export class Node {
    static getCluster(n: Node) {
        return n.guid.split('/')[0];
    }

    static getNamespace(n: Node) {
        return n.guid.split('/')[1];
    }

    static getKind(n: Node) {
        return n.guid.split('/')[2];
    }

    static getName(n: Node) {
        return n.guid.split('/')[3];
    }

    static getIcon(n: Node) {
        return Node.getKind(n)
            .substring(0, 1)
            .toUpperCase();
    }

    guid: string;
    label: string;
    phase: string;
}

export interface Graph {
    nodes?: Node[];
    edges?: Edge[];
}

export interface Edge {
    x: string;
    y: string;
}
