import { createStore, combineReducers, applyMiddleware } from 'redux';
import thunkMiddleware                  from 'redux-thunk';
import createLogger                     from 'redux-logger';
import { routeReducer }                 from 'redux-simple-router';
import { reducer as formReducer }       from 'redux-form';
import appReducers                      from '../reducers';


const reducers = combineReducers(Object.assign({}, appReducers, {
    routing: routeReducer,
    form:    formReducer
}));

const createStoreWithMiddleware = applyMiddleware(
    thunkMiddleware,
    createLogger()
)(createStore);


export default function configureStore(initialState) {
    const store = createStoreWithMiddleware(reducers, initialState);

    /*
    if (module.hot) {
        // Enable Webpack hot module replacement for reducers
        module.hot.accept('../reducers', () => {
            const nextRootReducer = require('../reducers');
            store.replaceReducer(nextRootReducer);
        });
    }
    */

    return store;
}
