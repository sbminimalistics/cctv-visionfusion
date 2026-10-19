import postgres from "postgres";

const DB_HOST = process.env.DB_HOST;
const DB_PORT = process.env.DB_PORT;
const DB_NAME = process.env.DB_NAME;
const DB_USER = process.env.DASHBOARD_DB_USER;
const DB_PASS = process.env.DASHBOARD_DB_PASS;

let sql: postgres.Sql = async function (input: string) {
    console.error("postgres client is not ready");
} as unknown as postgres.Sql;

const options: postgres.Options<{}> = {
    host: DB_HOST,
    port: parseInt(DB_PORT || ""),
    database: DB_NAME,
    username: DB_USER,
    password: DB_PASS,
};

if (
    Object.values(options).reduce((acc, val: string | number) => {
        console.log(val, val != null, val != "", val !== 0);
        return acc && val != null && val != "" && val !== 0;
    }, true) === false
) {
    console.error(
        "database connection options missing,",
        `values provided ${JSON.stringify(options)}`,
    );
} else {
    sql = postgres(options);
}

export default sql;
