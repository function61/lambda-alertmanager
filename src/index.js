var https = require('https')
  , AWS = require('aws-sdk')
  , utils = require('./utils');

var sns = new AWS.SNS();

var INGEST_TOPIC = process.env.INGEST_TOPIC;

if (!INGEST_TOPIC) {
	throw new Error("INGEST_TOPIC not defined");
}

var webCheckTimeoutMs = 8000;

var webChecks = [];

/*
	HOME: /root
	PWD: /app
	CHECK1: {"url":"https://example.com/"§"find":"Welcome to Example Ltd"}
	CHECK2: {"url":"http://blog.example.com/"§"find":"Welcome to our blog"}

	=> webChecks = [ { url: 'https://example.com/', find: 'Welcome to Example Ltd' }, { url: 'http://blog.example.com/', find: 'Welcome to our blog' } ]
*/
for (var key in process.env) {
	if (/^CHECK[0-9]+$/.test(key)) {
		/*	why § in JSON instead of ","? Well, apparently the geniuses at AWS decided to ship support for ENV variables
			with only not-supported character being "," (ENV variables probably serialized by "," at some subsystem - lazy programming?),
			effectively denying JSON and any list of values: https://forums.aws.amazon.com/thread.jspa?messageID=753580

			With comma we get this error: Member must satisfy regular expression pattern: [^,]*
		*/
		var awsLimitationStupidityDecoded = process.env[key].replace(/§/g, ',');

		webChecks.push(JSON.parse(awsLimitationStupidityDecoded));
	}
}

function checkWebProperty(checkDefinition, failOrSucceed) {
	var webClient = utils.getHttpOrHttpsWebClient(checkDefinition.url);

    var started = new Date().getTime();

    var resolved = false;

	var request = webClient.get(checkDefinition.url, function (response){
		var buffer = '';

		response.on('data', function (chunk){
			buffer += chunk.toString();
		});

		response.on('error', function (err){
			clearTimeout(timeout);
			if (!resolved) {
				resolved = true;
				failOrSucceed(err);
			}
		});

		response.on('end', function (){
			var err = null;

            var duration = new Date().getTime() - started;

			if (buffer.indexOf(checkDefinition.find) === -1) {
				err = new Error('find="' + checkDefinition.find + '" NOT in body: ' + buffer);
			}

			clearTimeout(timeout);
			if (!resolved) {
				resolved = true;
				failOrSucceed(err, duration);
			}
		});
	});

	request.on('error', function (err){
		clearTimeout(timeout);

		if (!resolved) {
			resolved = true;
			failOrSucceed(err);
		}
	});

	var timeout = null;

	timeout = setTimeout(function (){
		if (!resolved) {
			resolved = true;
			failOrSucceed(new Error("Timeout (" + webCheckTimeoutMs + "ms)"), webCheckTimeoutMs);
		}

		request.abort(); // triggers "error" event on "request"
	}, webCheckTimeoutMs);
}

// alert = { subject: ..., details: ... }
function postAlert(alert, next) {
	sns.publish({
		Message: alert.details,
		Subject: alert.subject,
		TopicArn: INGEST_TOPIC
	}, next);
}

function checkOne(checkDefinition, next) {
	if (!checkDefinition.url || !checkDefinition.find) {
		throw new Error('url or find not defined');
	}

	var resolved = false;

	function failOrSucceed(err, duration) {
		if (resolved) { // could be called many times
			return;
		}

		resolved = true;

		var suffix = err ? ' => ' + utils.oneLinerize(utils.truncate(err.message, 64)) : ' duration=' + duration;

		console.log((err ? '✗ ' : '✓ ') + checkDefinition.url + suffix);

		if (err) {
            postAlert({ subject: checkDefinition.url, details: err.message }, function (alertPostErr){
            	if (alertPostErr) { // alert posting failed - not much we can do :(
            		console.log('ALERT POSTING ERROR', alertPostErr);
            		next(alertPostErr); // TODO: wrap error with text that makes super clear what happened
            		return;
            	}

            	// alert posting succeeded (but actual check still failed)
            	next(err);
            });
		}
		else {
			next(null);
		}
	}

	checkWebProperty(checkDefinition, function (err, duration){
		if (err) { // re-try only once
			console.log(checkDefinition.url + ' failed once - re-trying (only once)');
			checkWebProperty(checkDefinition, failOrSucceed);
		}
		else {
			failOrSucceed(null, duration);
		}
	});
}

function launchOne(next) {
	var checkDefinition = this;

	checkOne(checkDefinition, next);
}

exports.handler = function(event, context) {
    console.log('Starting Canary. Check count: ' + webChecks.length);

    utils.parallel(webChecks.map(function (item){ return launchOne.bind(item); }), function (err, stats){
    	if (err) {
    		console.log('=> FAIL (' + stats.succeeded + '/' + stats.total + ') succeeded');
    	}
    	else {
    		console.log('=> All passed. Awesome!');
    	}
    	
    	context.succeed('Canary finished');
    });
};
