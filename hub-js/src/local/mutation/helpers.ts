import { UpdateTypeInstanceError } from "./update-type-instances";


/**
 * In Cypher we throw custom errors, e.g.:
 *   CALL apoc.util.validate(size(foundIDs) < size(allInputIDs), apoc.convert.toJson({code: 404, foundIDs: foundIDs}), null)
 * which produce such output:
 *   Failed to invoke procedure `apoc.cypher.doIt`: Caused by: java.lang.RuntimeException: {"lockedIDs":["b0283e96-ce83-451c-9325-0d144b9cea6a"],"code":409}
 * This functions tries to extract this error if possible.
 *
 * @param error - Cypher error.
 * @returns extracted error, if not possible, returns null.
 *
 */
export function tryToExtractCustomError(
  error: Error
): UpdateTypeInstanceError | null {
  const firstOpen = error.message.indexOf("{");
  const firstClose = error.message.lastIndexOf("}");
  const candidate = error.message.substring(firstOpen, firstClose + 1);

  try {
    return JSON.parse(candidate);
  } catch (e) {
    /* cannot extract, return generic error */
  }

  return null;
}
