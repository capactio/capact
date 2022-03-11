export interface TypeInstanceBackendInput {
  id: string;
  context?: unknown;
}

export interface CreateTypeInstanceInput {
  alias?: string;
  backend?: TypeInstanceBackendInput;
  value?: unknown;
}

export interface TypeInstanceBackendDetails {
  abstract: boolean;
  id: string;
  context?: unknown;
}

export interface CreateTypeInstancesInput {
  typeInstances: CreateTypeInstanceInput[];
  usesRelations: TypeInstanceUsesRelationInput[];
}

export interface TypeInstanceUsesRelationInput {
  from: string;
  to: string;
}
