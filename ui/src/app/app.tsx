import * as React from 'react';
import * as dagre from 'dagre';

const request = require('supera' +
    'gent');

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

export class App extends React.Component<{}, Graph> {

    constructor(props: Readonly<{}>) {
        super(props);
        this.state = {};
    }

    componentDidMount() {
        request.get("/api/graph")
            .then((r: { text: string }) => {
                this.setState(JSON.parse(r.text) as Graph);
            })
            .catch((e: Error) => console.log(e));
    }

    public render() {
        const g = new dagre.graphlib.Graph();
        g.setGraph({rankdir: "LR"});
        g.setDefaultEdgeLabel(function () {
            return {};
        });
        (this.state.vertices || []).forEach(v => g.setNode(v, {label: v, width: 40, height: 20}));
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
            <div className='graph'>
                {g.nodes().map((id) => g.node(id)).map((n) => <div key={n.label} style={{
                    position: "absolute",
                    left: n.x - n.width / 2,
                    top: n.y - n.height / 2,
                    width: n.width,
                    height: n.height,
                    lineHeight: n.height + "px",
                    borderRadius: n.width / 2,
                    textAlign: "center"
                }}>{n.label}</div>)}
                {edges.map(edge => (
                    <div key={`${edge.from}-${edge.to}`}>
                        {edge.lines.map((line, i) => {
                            const distance = Math.sqrt(Math.pow(line.x1 - line.x2, 2) + Math.pow(line.y1 - line.y2, 2));
                            const xMid = (line.x1 + line.x2) / 2;
                            const yMid = (line.y1 + line.y2) / 2;
                            const angle = (Math.atan2(line.y1 - line.y2, line.x1 - line.x2) * 180) / Math.PI;
                            return (
                                <div
                                    key={i}
                                    style={{
                                        position: "absolute",
                                        width: distance,
                                        left: xMid - distance / 2,
                                        top: yMid,
                                        transform: `rotate(${angle}deg)`,
                                        borderTop: "1px solid #ddd"
                                    }}
                                />
                            );
                        })}
                    </div>
                ))}
            </div>
        );
    }

}
