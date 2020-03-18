import * as React from 'react';
import {Node} from "./types";

const request = require('superagent');

interface State {
    node?: Node;
}

export class NodeInfoPanel extends React.Component<{ guid: string }, State> {
    constructor(props: Readonly<{ guid: string }>) {
        super(props);
        this.state = {};
    }

    componentDidMount() {
        request
            .get('/api/v1/nodes/' + this.props.guid)
            .then((r: { text: string }) => this.setState({node: JSON.parse(r.text) as Node}))
            .catch((e: Error) => console.log(e));
    }

    public render() {
        return (
            <div>
                <h4>{this.state.node && this.state.node.label}</h4>
            </div>
        );
    }
}
