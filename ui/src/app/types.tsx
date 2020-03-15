export class Vertex {
    static getKind: (v: Vertex) => string;
    guid: string;
    label: string;
}

Vertex.getKind = (v: Vertex) => {
    return v.guid.split("/")[2];
};

export interface Graph {
    vertices?: Vertex[];
    edges?: Edge[];
}

export interface Edge {
    x: string;
    y: string;
}
