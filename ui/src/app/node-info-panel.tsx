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
                {this.state.node && (<>
                    <h4>{this.state.node.label}</h4>
                    <div className='white-box'>
                        <div className='row'>
                            <div className='columns small-3'>CLUSTER</div>
                            <div className='columns small-9'>{Node.getCluster(this.state.node)}</div>
                        </div>
                        <div className='row'>
                            <div className='columns small-3'>NAMESPACE</div>
                            <div className='columns small-9'>{Node.getNamespace(this.state.node)}</div>
                        </div>
                        <div className='row'>
                            <div className='columns small-3'>KIND</div>
                            <div className='columns small-9'>{Node.getKind(this.state.node)}</div>
                        </div>
                        <div className='row'>
                            <div className='columns small-3'>NAME</div>
                            <div className='columns small-9'>{Node.getName(this.state.node)}</div>
                        </div>
                    </div>
                </>)}
            </div>
        );
    }
}
