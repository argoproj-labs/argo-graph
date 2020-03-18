import * as React from 'react';
import * as dagre from 'dagre';
import {Graph, Node} from './types';
import {ZeroState} from "./zero-state";

const request = require('superagent');
require('./graph-panel.scss');

interface Line {
    x1: number;
    y1: number;
    x2: number;
    y2: number;
}

interface Props {
    guid: string;
    onSelect: (guid: string) => void;
}

interface State {
    graph: Graph;
}

export class GraphPanel extends React.Component<Props, State> {
    private graphBox: React.RefObject<any>;
    private width: number;

    constructor(props: Readonly<Props>) {
        super(props);
        this.state = {graph: {}};
        this.graphBox = React.createRef()
    }

    componentDidMount() {
        this.width = this.graphBox.current.offsetWidth;
        request
            .get('/api/v1/graph/' + this.props.guid)
            .then((r: { text: string }) => this.setState({graph: JSON.parse(r.text) as Graph}))
            .catch((e: Error) => console.log(e));
    }

    public render() {
        return <div ref={this.graphBox}>
            {!this.state.graph.nodes && <ZeroState title='No nodes'/> || this.renderGraph()}
        </div>
    }

    renderGraph() {
        const nodeSize = 48;
        const ranksep = 100;
        const g = new dagre.graphlib.Graph();
        g.setGraph({rankdir: 'RL', ranksep: ranksep});
        g.setDefaultEdgeLabel(() => ({}));
        (this.state.graph.nodes || []).forEach(v =>
            g.setNode(v.guid, {
                icon: Node.getIcon(v),
                label: v.label,
                width: nodeSize,
                height: nodeSize
            })
        );
        (this.state.graph.edges || []).forEach(e => g.setEdge(e.x, e.y));

        dagre.layout(g);
        const edges: { from: string; to: string; lines: Line[] }[] = [];
        g.edges().forEach(v => {
            const edge = g.edge(v);
            const lines: Line[] = [];
            if (edge.points.length > 1) {
                for (let i = 1; i < edge.points.length; i++) {
                    lines.push({
                        x1: edge.points[i - 1].x,
                        y1: edge.points[i - 1].y,
                        x2: edge.points[i].x,
                        y2: edge.points[i].y
                    });
                }
            }
            edges.push({from: v.v, to: v.w, lines});
        });

        const nodes = g.nodes().map(id => ({...g.node(id), ...{id: id}}));
        const width = nodes.map(n => n.x + n.width + nodeSize).reduce((l, r) => Math.max(l, r), 0);
        const height = nodes.map(n => n.y + n.height + nodeSize).reduce((l, r) => Math.max(l, r), 0);
        const left = (this.width - width) / 2 + nodeSize;
        const top = nodeSize * 2;

        return (<>
                <div className='graph' style={{width: width, height: height}}>
                    {nodes.map((n: any) => (
                        <>
                            <div
                                key={`node-${n.id}`}
                                className='node'
                                onClick={() => this.props.onSelect(n.id)}
                                style={{
                                    position: 'absolute',
                                    left: left + n.x - nodeSize / 2,
                                    top: top + n.y - nodeSize / 2,
                                    width: nodeSize,
                                    height: nodeSize,
                                    borderRadius: nodeSize / 2,
                                    textAlign: 'center',
                                    lineHeight: nodeSize + 'px',
                                    fontWeight: n.id == this.props.guid ? 'bold' : 'normal'
                                }}>
                                {n.icon}
                            </div>
                            <div
                                key={`label-${n.label}`}
                                style={{
                                    position: 'absolute',
                                    left: left + n.x - ranksep / 2,
                                    top: top + n.y + nodeSize / 2,
                                    width: ranksep,
                                    textAlign: 'center',
                                    fontSize: '0.75em',
                                    overflow: 'hidden',
                                    textOverflow: 'ellipsis'
                                }}>
                                {n.label}
                            </div>
                        </>
                    ))}
                    {edges.map(edge => (
                        <div key={`edge-${edge.from}-${edge.to}`}>
                            {edge.lines.map((line, i) => {
                                const distance = Math.sqrt(Math.pow(line.x1 - line.x2, 2) + Math.pow(line.y1 - line.y2, 2));
                                const xMid = (line.x1 + line.x2) / 2;
                                const yMid = (line.y1 + line.y2) / 2;
                                const angle = (Math.atan2(line.y1 - line.y2, line.x1 - line.x2) * 180) / Math.PI;
                                return (
                                    <div
                                        key={`line-${edge.from}-${edge.to}-${i}`}
                                        className='line'
                                        style={{
                                            position: 'absolute',
                                            width: distance,
                                            left: left + xMid - distance / 2,
                                            top: top + yMid,
                                            transform: `rotate(${angle}deg)`,
                                            borderTop: '1px solid #888'
                                        }}
                                    />
                                );
                            })}
                        </div>
                    ))}
                </div>
            </>
        );
    }
}
