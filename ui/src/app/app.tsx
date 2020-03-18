import * as PropTypes from 'prop-types';
import * as React from 'react';
import {Layout} from 'argo-ui';
import {GraphPage} from './graph-page';
import {Redirect, Route, Router, Switch, useParams} from 'react-router-dom';
import {createBrowserHistory} from 'history';
import {NodeListPage} from './node-list-page';

export const history = createBrowserHistory();
export const {Provider} = React.createContext(null);

export class App extends React.Component<{}, {}> {
    public static childContextTypes = {
        history: PropTypes.object,
        router: PropTypes.object
    };

    constructor(props: Readonly<{}>) {
        super(props);
    }

    public render() {
        return (
            <Provider value={this.context}>
                <Router history={history}>
                    <Layout navItems={[{title: 'Nodes', iconClassName: 'fa fa-list', path: '/nodes'}]}>
                        <Switch>
                            <Route path='/nodes'>
                                <NodeListPage />
                            </Route>
                            <Route
                                path='/graph/:cluster/:namespace/:kind/:name'
                                component={() => {
                                    const {cluster, namespace, kind, name} = useParams();
                                    return <GraphPage guid={cluster + '/' + namespace + '/' + kind + '/' + name} />;
                                }}
                            />
                            <Route>
                                <Redirect to='/nodes' />
                            </Route>
                        </Switch>
                    </Layout>
                </Router>
            </Provider>
        );
    }

    public getChildContext() {
        return {
            history: history,
            router: {route: {location: {pathname: '/nodes'}}, history: history}
        };
    }
}
