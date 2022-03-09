export enum CustomCypherErrorCode {
  BadRequest = 400,
  Conflict = 409,
  NotFound = 404,
}

export interface CustomCypherErrorOutput {
  code: CustomCypherErrorCode;

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  [key: string]: any;
}

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
export function tryToExtractCustomCypherError(
  error: Error
): CustomCypherErrorOutput | null {
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
