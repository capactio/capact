const getSession = (context) => context.driver.session();

export async function runSingleQuery(context, query: string, cypherParams) {
  const session = getSession(context);

  let result: any;

  try {
    result = await session.readTransaction(async (tx) => {
      const res = await tx.run(query, cypherParams);
      return res.records;
    });
  } finally {
    session.close();
  }
  return result;
}

export async function runSingleMutation(context, query: string, cypherParams) {
  const session = getSession(context);

  let result: any;

  try {
    result = await session.writeTransaction(async (tx) => {
      const res = await tx.run(query, cypherParams);
      return res.records;
    });
  } finally {
    session.close();
  }
  return result;
}

export default {
};
