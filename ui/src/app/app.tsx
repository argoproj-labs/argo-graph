import * as PropTypes from 'prop-types';
import * as React from 'react';
import {Layout} from 'argo-ui';
import {GraphPage} from "./graph-page";
import {Route, Router,} from 'react-router-dom';
import {createBrowserHistory} from "history";

export const history = createBrowserHistory();
export const {Provider} = React.createContext(null);

export class App extends React.Component<{}, {}> {
    public static childContextTypes = {
        history: PropTypes.object,
        router: PropTypes.object,
    };

    constructor(props: Readonly<{}>) {
        super(props);
    }

    public render() {
        return (
            <Provider value={this.context}>
                <Router history={history}>
                    <Layout navItems={[
                        {title: "Home", iconClassName: 'fa fa-project-diagram', path: ""}
                    ]}>
                        <Route component={GraphPage}/>
                    </Layout>
                </Router>
            </Provider>
        );
    }

    public getChildContext() {
        return {
            history: history,
            router: {route: {location: {pathname: ""}}}
        }
    }
}
