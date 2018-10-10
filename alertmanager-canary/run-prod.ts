import { ProdActions } from './actions';
import { handlerInternal } from './index';

// this exists pretty much for so you can simulate production by running:
//     $ node run-prod.js

handlerInternal(new ProdActions()).then(noopForLinter, noopForLinter);

function noopForLinter() {
	/* noop */
}
