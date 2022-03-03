export interface TypeInstanceBackendInput {
  id: string;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  context?: any;
}

export interface CreateTypeInstanceInput {
  alias?: string;
  backend?: TypeInstanceBackendInput;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  value?: any;
}

export interface TypeInstanceBackendDetails {
  abstract: boolean;
  id: string;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  context?: any;
}

export interface CreateTypeInstancesInput {
  typeInstances: CreateTypeInstanceInput[];
  usesRelations: TypeInstanceUsesRelationInput[];
}

export interface TypeInstanceUsesRelationInput {
  from: string;
  to: string;
}
