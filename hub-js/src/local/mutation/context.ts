import { Driver } from "neo4j-driver";
import DelegatedStorageService from "../storage/service";

export interface ContextWithDriver {
  driver: Driver;
}

export interface ContextWithDelegatedStorage {
  delegatedStorage: DelegatedStorageService;
}

export interface Context extends ContextWithDriver, ContextWithDelegatedStorage {
}
