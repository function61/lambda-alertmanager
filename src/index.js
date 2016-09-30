var https = require('https')
  , utils = require('./utils');

var webCheckTimeoutMs = 8000;

var webChecks = [
	{
		url: 'http://example.com/',
		findString: 'string to find'
	}
];

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

			if (buffer.indexOf(checkDefinition.findString) === -1) {
				err = new Error('findString="' + checkDefinition.findString + '" NOT in body: ' + buffer);
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

function postAlert(alertBody, next) {
	var request = https.request({
			hostname: 'b4eac0iqek.execute-api.us-east-1.amazonaws.com',
			method: 'POST',
			path: '/prod/ingest',
			headers: {
				'Content-Type': 'application/json',
				'Content-Length': Buffer.byteLength(alertBody)
			},
		}, function (response){
		response.on('error', next);
		response.on('data', function (chunk){
			console.log('Alert ingestor response: ' + chunk.toString());
		});

		response.on('end', function(){
			next(null);
		});
	});

	request.on('error', next);
	request.end(alertBody);
}

function checkOne(checkDefinition, next) {
	if (!checkDefinition.url || !checkDefinition.findString) {
		throw new Error('url or findString not defined');
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
			var alertBody = JSON.stringify({
				subject: checkDefinition.url,
				details: err.message
			});

            postAlert(alertBody, function (alertPostErr){
                next(alertPostErr || err);
            });
		}
		else {
			next(err);
		}
	}

	checkWebProperty(checkDefinition, function (err, duration){
		if (err) { // re-try only once
			console.log(checkDefinition.url + ' failed once - re-trying (only once)');
			checkWebProperty(checkDefinition, failOrSucceed);
		}
		else {
			failOrSucceed(err, duration);
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
