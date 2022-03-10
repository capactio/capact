export enum Operation {
  None,
  UpdateTypeInstancesMutation,
}

export default class UpdateArgsContainer {
  private valuesPerTypeInstance: Map<string, unknown>;
  private latestKnownRevPerTypeInstance: Map<string, number>;
  private currentOperation: Operation;
  private currentOperationOwnerID: Map<string, string | undefined>;

  constructor() {
    this.valuesPerTypeInstance = new Map();
    this.latestKnownRevPerTypeInstance = new Map();
    this.currentOperationOwnerID = new Map();
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
  SetValue(id: string, value: unknown) {
    return this.valuesPerTypeInstance.set(id, value);
  }

  /**
   * Gives an option to fetch the input set by other resolver.
   *
   *
   * @return - TypeInstance's value if set by other resolver.
   *
   */
  GetValue(id: string): unknown {
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
   * Used to optimize number of request. If already stored, we don't need to trigger the update logic.
   *
   * @param id - TypeInstance's ID.
   * @return - - TypeInstance's last revision version. If not set, returns 0.
   * It's a safe assumption as Local Hub starts counting revision with 1.
   *
   */
  GetLastKnownRev(id: string) {
    return this.latestKnownRevPerTypeInstance.get(id) || 0;
  }

  /**
   * Returns the owner id of the current operation (sent in GQL request).
   *
   * @return - current owner id.
   *
   */
  GetOwnerID(id: string): string | undefined {
    return this.currentOperationOwnerID.get(id);
  }

  /**
   * Set the owner id of the current operation (sent in GQL request). May be undefined.
   *
   * @param id - related TypeInstance id.
   * @param ownerId - current owner id.
   *
   */
  SetOwnerID(id: string, ownerId?: string) {
    this.currentOperationOwnerID.set(id, ownerId);
  }
}
