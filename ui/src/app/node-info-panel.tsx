import * as React from 'react';


export class NodeInfoPanel extends React.Component<{ guid: string }, {}> {
    constructor(props: Readonly<{ guid: string }>) {
        super(props);
        this.state = {};
    }

    public render() {
        return (
            <div>

                {this.props.guid}
            </div>
        );
    }
}
