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

  /**
   * Sets which operation started the GraphQL request.
   * It is used to correlate proper flow between different resolvers.
   *
   *
   * @param op - GraphQL request operation.
   *
   */
  SetOperation(op: Operation) {
    if (this.currentOperation == op) {
      return;
    }

    if (this.currentOperation != Operation.None) {
      throw Error(`Operation in progress, cannot change it`);
    }

    this.currentOperation = op;
  }

  /**
   * Describes which operation is currently in progress.
   *
   *
   * @return - GraphQL request operation.
   *
   */
  GetOperation(): Operation {
    return this.currentOperation;
  }

  /**
   * Gives an option to transmit the TypeInstance's input value between resolvers.
   *
   *
   * @param id - TypeInstance's ID.
   * @param value - User specified TypeInstance's value.
   *
   */
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  SetValue(id: string, value: any) {
    return this.valuesPerTypeInstance.set(id, value);
  }

  /**
   * Gives an option to fetch the input set by other resolver.
   *
   *
   * @return - TypeInstance's value if set by other resolver.
   *
   */
  GetValue(id: string): GetValueOutput {
    return this.valuesPerTypeInstance.get(id);
  }

  /**
   * Informs which TypeInstance's version was already processed a stored.
   *
   *
   * @param id - TypeInstance's ID.
   * @param rev - TypeInstance's last revision version.
   *
   */
  SetLastKnownRev(id: string, rev: number) {
    this.latestKnownRevPerTypeInstance.set(id, rev);
  }

  /**
   * Returns latest TypeInstance's revision version.
   * Used to optimize number of request. If already stored, we don't need to trigger the udpate logic.
   *
   * @param id - TypeInstance's ID.
   * @return - - TypeInstance's last revision version. If not set, returns 0.
   * It's a safe assumption as Local Hub starts counting revision with 1.
   *
   */
  GetLastKnownRev(id: string) {
    return this.latestKnownRevPerTypeInstance.get(id) || 0;
  }
}
