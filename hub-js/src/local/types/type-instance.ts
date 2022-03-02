export interface TypeInstanceBackendInput {
  id: string;
  context?: any;
}

export interface CreateTypeInstanceInput {
  alias?: string;
  backend?: TypeInstanceBackendInput;
  value?: any;
}

export interface TypeInstanceBackendDetails {
  abstract: boolean;
  id: string;
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
