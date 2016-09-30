require('./index').handler(null, {
	succeed: function (){
		console.log('context.succeed()', arguments);
	},

	fail: function (){
		console.log('context.fail()', arguments);
	}
});
