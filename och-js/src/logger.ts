import { createLogger, format, transports } from "winston";

export const logger = createLogger({
  level: "info",
  format: format.combine(
    format.timestamp(),
    format.printf(({ timestamp, level, message, ...fields }) => {
      let fmt = `${timestamp}\t${level.toUpperCase()}\t${message}`;
      if (Object.keys(fields).length > 0) {
        fmt += `\t${JSON.stringify(fields)}`;
      }

      return fmt;
    })
  ),
  transports: [new transports.Console()],
});
