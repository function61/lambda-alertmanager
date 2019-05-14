import { ProdActions } from './actions';
import { handlerWithActions } from './index';
import { mockScheduledEvent } from './tests';

// this exists pretty much for so you can simulate production by running:
//     $ node simulate-prod.js

handlerWithActions(mockScheduledEvent(), new ProdActions()).then(noopForLinter, noopForLinter);

function noopForLinter() {
	/* noop */
}
