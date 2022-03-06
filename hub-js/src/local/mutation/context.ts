import { Driver } from "neo4j-driver";
import DelegatedStorageService from "../storage/service";
import UpdateArgsContext from "../storage/update-args-context";

export interface ContextWithDriver {
  driver: Driver;
}

export interface ContextWithDelegatedStorage {
  delegatedStorage: DelegatedStorageService;
}

export interface ContextWithUpdateArgs {
  updateArgs: UpdateArgsContext;
}

export interface Context
  extends ContextWithDriver,
    ContextWithDelegatedStorage,
    ContextWithUpdateArgs {}
