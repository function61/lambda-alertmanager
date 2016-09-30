var http = require('http')
  , https = require('https');

function oneLinerize(str) {
	return str.replace(/\n/g, '\\n');
}

function truncate(str, to) {
	if (to >= str.length) {
		return str;
	}

	return str.substr(0, to - 2) + '..';
}

// FFS JavaScript how's it so hard to do this?
function parallel(tasks, next) {
	var succeeded = 0;
	var failed = 0;

	function _parallelSingle(task, next) {
		var resolved = false;

		task(function (err){
			if (resolved) {
				throw new Error('Cannot resolve twice');
			}

			resolved = true;

			if (err) {
				next(err);
			}
			else {
				next(null);
			}
		});
	}

	function _parallelSingle_result(err) {
		if (err) {
			failed++;
		}
		else {
			succeeded++;
		}

		if ((failed + succeeded) === tasks.length) {
			next(failed !== 0, { succeeded: succeeded, failed: failed, total: succeeded + failed });
		}
	}

	for (var i = 0; i < tasks.length; ++i) {
		_parallelSingle(tasks[i], _parallelSingle_result);
	}
}

// FFS node.js why do I have to write this wrapper?!
var getHttpOrHttpsWebClient = function (url) {
	if (/^http:/.test(url)) {
		return http;
	}

	if (/^https:/.test(url)) {
		return https;
	}

	throw new Error('Unknown protocol in URL: ' + url);
};

module.exports = {
	oneLinerize: oneLinerize,
	truncate: truncate,
	parallel: parallel,
	getHttpOrHttpsWebClient: getHttpOrHttpsWebClient
};
