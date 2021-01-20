const getSession = (context: any) => context.driver.session();

export async function runSingleQuery(
  context: any,
  query: string,
  cypherParams: any
) {
  const session = getSession(context);

  let result: any;

  try {
    result = await session.readTransaction(async (tx: any) => {
      const res = await tx.run(query, cypherParams);
      return res.records;
    });
  } finally {
    session.close();
  }
  return result;
}

export async function runSingleMutation(
  context: any,
  query: string,
  cypherParams: any
) {
  const session = getSession(context);

  let result: any;

  try {
    result = await session.writeTransaction(async (tx: any) => {
      const res = await tx.run(query, cypherParams);
      return res.records;
    });
  } finally {
    session.close();
  }
  return result;
}

export default {};
