import { Driver } from "neo4j-driver";
import DelegatedStorageService from "../../storage/service";
import UpdateArgsContainer from "../../storage/update-args-container";

export interface ContextWithDriver {
  driver: Driver;
}

export interface ContextWithDelegatedStorage {
  delegatedStorage: DelegatedStorageService;
}

export interface ContextWithUpdateArgs {
  updateArgs: UpdateArgsContainer;
}

export interface Context
  extends ContextWithDriver,
    ContextWithDelegatedStorage,
    ContextWithUpdateArgs {}
