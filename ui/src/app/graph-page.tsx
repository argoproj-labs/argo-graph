import * as React from 'react';
import * as dagre from 'dagre';
import {Page} from "argo-ui/src/index";

const request = require('superagent');

interface Graph {
    vertices?: string[];
    edges?: Edge[];
}

interface Edge {
    x: string;
    y: string;
}

interface Line {
    x1: number;
    y1: number;
    x2: number;
    y2: number;
}

export class GraphPage extends React.Component<{}, Graph> {

    constructor(props: Readonly<{}>) {
        super(props);
        this.state = {};
    }

    componentDidMount() {
        request.get("/api/v1/graph")
            .then((r: { text: string }) => {
                this.setState(JSON.parse(r.text) as Graph);
            })
            .catch((e: Error) => console.log(e));
    }

    public render() {
        const vertexSize = 32;
        const ranksep = 100;
        const g = new dagre.graphlib.Graph();
        g.setGraph({rankdir: "LR", "ranksep": ranksep});
        g.setDefaultEdgeLabel(() => ({}));
        (this.state.vertices || []).forEach(v => g.setNode(v, {label: v, width: vertexSize, height: vertexSize}));
        (this.state.edges || []).forEach(e => g.setEdge(e.x, e.y));
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
                        y2: edge.points[i].y,
                    });
                }
            }
            edges.push({from: v.v, to: v.w, lines});
        });

        return (
            <Page title='Argo Graph'>
                <div className='row'>
                    <div className='columns small-12'>
                        <div className='graph' style={{margin: 40}}>
                            {g.nodes().map((id) => g.node(id)).map((n) => <>
                                <div key={`vertex-${n.label}`} style={{
                                    position: "absolute",
                                    left: n.x - vertexSize / 2,
                                    top: n.y - vertexSize / 2,
                                    width: vertexSize,
                                    height: vertexSize,
                                    borderRadius: vertexSize / 2,
                                    backgroundColor: "purple",
                                    border: "1px solid #888"
                                }}/>
                                <div key={`label-${n.label}`} style={{
                                    position: "absolute",
                                    left: n.x - ranksep / 2,
                                    top: n.y + vertexSize / 2,
                                    width: ranksep,
                                    textAlign: "center",
                                    fontSize: "0.75em",
                                    overflow: "hidden",
                                    textOverflow: "ellipsis"
                                }}>{n.label}</div>
                            </>)}
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
                                                style={{
                                                    position: "absolute",
                                                    width: distance,
                                                    left: xMid - distance / 2,
                                                    top: yMid,
                                                    transform: `rotate(${angle}deg)`,
                                                    borderTop: "1px solid #888"
                                                }}
                                            />
                                        );
                                    })}
                                </div>
                            ))}
                        </div>
                    </div>
                </div>
            </Page>
        );
    }

}
