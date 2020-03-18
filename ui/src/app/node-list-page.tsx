import * as React from 'react';
import {Page} from 'argo-ui/src/index';
import {Node} from './types';
import {ZeroState} from './zero-state';
import {PhaseIcon} from './phase-icon';

const request = require('superagent');

export class NodeListPage extends React.Component<{}, {nodes?: Node[]}> {
    constructor(props: Readonly<{}>) {
        super(props);
        this.state = {};
    }

    componentDidMount() {
        request
            .get('/api/v1/nodes')
            .then((r: {text: string}) => this.setState({nodes: JSON.parse(r.text) as Node[]}))
            .catch((e: Error) => console.log(e));
    }

    public render() {
        return (
            <Page title='Nodes' toolbar={{breadcrumbs: [{title: 'Nodes', path: '/nodes'}]}}>
                {(this.state.nodes && this.state.nodes.length > 0 && (
                    <>
                        <div className='argo-table-list'>
                            <div className='row argo-table-list__head'>
                                <div className='columns large-1'>PHASE</div>
                                <div className='columns large-3'>LABEL</div>
                                <div className='columns large-2'>CLUSTER</div>
                                <div className='columns large-2'>NAMESPACE</div>
                                <div className='columns large-2'>KIND</div>
                                <div className='columns large-2'>NAME</div>
                            </div>
                            {this.state.nodes.map(node => (
                                <div className='row row argo-table-list__row'>
                                    <div className='columns large-1'>
                                        <PhaseIcon phase={node.phase} />
                                    </div>
                                    <div className='columns large-3'>
                                        <a href={'/graph/' + node.guid}>{node.label || node.guid}</a>
                                    </div>
                                    <div className='columns large-2'>{Node.getCluster(node)} </div>
                                    <div className='columns large-2'>{Node.getNamespace(node)} </div>
                                    <div className='columns large-2'>{Node.getKind(node)} </div>
                                    <div className='columns large-2'>{Node.getName(node)} </div>
                                </div>
                            ))}
                        </div>
                    </>
                )) || <ZeroState title='No nodes' />}
            </Page>
        );
    }
}
