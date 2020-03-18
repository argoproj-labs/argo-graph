import * as React from 'react';
import {Page, SlidingPanel} from 'argo-ui/src/index';
import {GraphPanel} from "./graph-panel";
import {NodeInfoPanel} from "./node-info-panel";

interface Props {
    guid: string;
}

interface State {
    selectedGuid?: string;
}

export class GraphPage extends React.Component<Props, State> {
    constructor(props: Readonly<Props>) {
        super(props);
        this.state = {};
    }

    public render() {
        return (
            <Page title='Graph' toolbar={{breadcrumbs: [{title: this.props.guid}]}}>
                <GraphPanel guid={this.props.guid} onSelect={selectedGuid => this.setState({selectedGuid})}/>
                {this.state.selectedGuid &&
                <SlidingPanel isShown={true} onClose={() => this.setState({selectedGuid: null})}>
                    <NodeInfoPanel guid={this.state.selectedGuid}/>
                </SlidingPanel>
                }
            </Page>
        );
    }
}
