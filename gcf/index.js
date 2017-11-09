const https = require('https')
const util 	= require('util')

/**
 * Triggered from a message on a Cloud Pub/Sub topic.
 *
 * @param {!Object} event The Cloud Functions event.
 * @param {!Function} The callback function.
 */
exports.updateSubscriber = function updateSubscriber (event, callback) {
  // The Cloud Pub/Sub Message object.
  //console.log(util.inspect(event))
  const slackMessage = composeMessage(event)
 // console.log("Sending: " + slackMessage)
  
  if (slackMessage) {
  	makeRequest(slackMessage)
  }
  //console.log(`My Cloud Function: ${event.data.message}`);  
  callback();
};

function composeMessage(event) {
	const pubsubEvent = event.data;
	const action = Buffer.from(pubsubEvent.data, 'base64').toString();
	var message = ""
	switch(action) {
		case "add": 
			message += "%s: %s from %s, %s just signed up!"
			break
		case "update":
			message += "%s: %s from %s, %s just updated (OKed the popup)!"
			break
			
	}
	if (! message) {
		return undefined
	}
	var p = pubsubEvent.attributes
	var country = p.hasOwnProperty("country") 
	    ? p["country"] 
	    : p["g_country"]
	var email = (p.hasOwnProperty("name")) ? util.format("%s <%s>", p["name"],p["email"]) : p["email"];
	var city = p["g_city"].replace(/\b([a-z]{1})/g, function(match) { return match.toUpperCase() } )
	var now = new Date().toISOString().replace(/T/, ' ').replace(/\..+/, '') + " UTC"
	message = util.format(message, now, email, city, country)
	return message
}

function makeRequest(message) {
	var postData = JSON.stringify({
		text: message
	});
	
	var options = {
		hostname: 'hooks.slack.com',
		path:'/services/!redacted!',
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
			'Content-Length': Buffer.byteLength(postData)
		}
	};
	
	options.agent = new https.Agent(options)
	
	const req = https.request(options, (res) => {
		console.log('statusCode:', res.statusCode);
  		console.log('headers:', res.headers);
  		res.on('data', (d) => {
    		console.log("Response:" + d);
    	});
  	});
	req.on('error', (e) => {
  		console.error(e);
	});
	
	req.write(postData);
	req.end();
}
