import * as React from 'react';

require('./phase-icon.scss');

export class PhaseIcon extends React.Component<{phase: string}, {}> {
    render() {
        const icon: string =
            ({
                Error: 'fa-exclamation-circle error',
                Failed: 'fa-cross-circle failed',
                Running: 'fa-circle-notch fa-spin running',
                Pending: 'fa-spinner fa-spin pending',
                Succeeded: 'fa-check-circle succeeded'
            } as any)[this.props.phase] || 'fa-dot-circle unknown';

        return <i className={'phase-icon fa ' + icon} title={this.props.phase} />;
    }
}
