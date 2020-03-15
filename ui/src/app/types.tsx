export class Node {
    static getKind: (v: Node) => string;
    guid: string;
    label: string;
}

Node.getKind = (v: Node) => {
    return v.guid.split("/")[2];
};

export interface Graph {
    nodes?: Node[];
    edges?: Edge[];
}

export interface Edge {
    x: string;
    y: string;
}
