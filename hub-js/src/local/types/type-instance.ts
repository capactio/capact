export interface TypeInstanceBackendInput {
  id: string;
  context?: undefined;
}

export interface CreateTypeInstanceInput {
  alias?: string;
  backend?: TypeInstanceBackendInput;
  value?: undefined;
}

export interface TypeInstanceBackendDetails {
  abstract: boolean;
  id: string;
  context?: undefined;
}

export interface CreateTypeInstancesInput {
  typeInstances: CreateTypeInstanceInput[];
  usesRelations: TypeInstanceUsesRelationInput[];
}

export interface TypeInstanceUsesRelationInput {
  from: string;
  to: string;
}
