import nodemailer from "nodemailer";
let transport: ReturnType<typeof nodemailer.createTransport> | undefined;
export async function sendAuthEmail(to: string, subject: string, text: string): Promise<void> {
  if (!process.env.SMTP_HOST || !process.env.EMAIL_FROM) throw new Error("SMTP_HOST and EMAIL_FROM are required");
  transport ??= nodemailer.createTransport({
    host: process.env.SMTP_HOST, port: Number(process.env.SMTP_PORT || "587"),
    secure: process.env.SMTP_SECURE === "true",
    requireTLS: Boolean(process.env.SMTP_USER) && process.env.SMTP_SECURE !== "true",
    ...(process.env.SMTP_USER ? { auth: { user: process.env.SMTP_USER, pass: process.env.SMTP_PASSWORD } } : {}),
    connectionTimeout: 10000, greetingTimeout: 10000, socketTimeout: 15000,
  });
  const result = await transport.sendMail({ from: process.env.EMAIL_FROM, to, subject, text });
  if (!result.accepted?.length || result.rejected?.length) throw new Error("Email was not accepted by the SMTP server");
}
