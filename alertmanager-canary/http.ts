import * as http from 'http';
import * as https from 'https';
import * as url from 'url';

export function httpGetBody(remoteUrl: string): Promise<string> {
	return new Promise<string>((resolve, reject) => {
		// childish node API allows us to either specify http.get(url) or http.get(options)
		// but not both. if you give options, you've to pass the different URL components
		// yourself.. what could go wrong......
		const urlParsed = new url.URL(remoteUrl);

		const req = getHttpOrHttps(
			{
				protocol: urlParsed.protocol,
				host: urlParsed.host,
				path: urlParsed.pathname + urlParsed.search,
			},
			(res) => {
				let buffer = '';

				res.on('error', (err) => {
					reject(err);
				});

				res.on('data', (chunk: Buffer) => {
					buffer += chunk.toString();
				});

				res.on('end', () => {
					resolve(buffer);
				});
			},
		);

		// https://stackoverflow.com/a/11221332
		req.setTimeout(8000, () => {
			req.abort();
		});

		// TODO: if request errors, can response error too?
		req.on('error', (err) => {
			reject(err);
		});
	});
}

// moronic node API forces us to call different functions for http/https URLs
function getHttpOrHttps(
	options: http.RequestOptions,
	callback: (res: http.IncomingMessage) => void,
): http.ClientRequest {
	if (options.protocol === 'http:') {
		return http.get(options, callback);
	} else if (options.protocol === 'https:') {
		return https.get(options, callback);
	}

	throw new Error(`Unknown protocol in URL: ${url}`);
}
