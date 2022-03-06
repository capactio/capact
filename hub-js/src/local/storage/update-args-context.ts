export enum Operation {
  None,
  UpdateTypeInstancesMutation,
}

interface GetValueOutput {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  value: any;
  latestKnownRevision: number;
}

export default class UpdateArgsContext {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  private valuesPerTypeInstance: Map<string, any>;
  private latestKnownRevPerTypeInstance: Map<string, number>;
  private currentOperation: Operation;

  constructor() {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    this.valuesPerTypeInstance = new Map<string, any>();
    this.latestKnownRevPerTypeInstance = new Map<string, number>();
    this.currentOperation = Operation.None;
  }

  SetOperation(op: Operation) {
    if (this.currentOperation == op) {
      return;
    }

    if (this.currentOperation != Operation.None) {
      throw Error(`Operation in progress, cannot change it`);
    }

    this.currentOperation = op;
  }

  GetOperation(): Operation {
    return this.currentOperation;
  }

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  SetValue(id: string, value: any) {
    return this.valuesPerTypeInstance.set(id, value);
  }

  GetValue(id: string): GetValueOutput {
    return this.valuesPerTypeInstance.get(id);
  }

  SetLastKnownRev(id: string, rev: number) {
    this.latestKnownRevPerTypeInstance.set(id, rev);
  }

  GetLastKnownRev(id: string) {
    return this.latestKnownRevPerTypeInstance.get(id) || 0;
  }
}
