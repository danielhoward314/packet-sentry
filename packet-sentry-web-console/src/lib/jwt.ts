export function parseJwt(token: string) {
  if (!token) return {};

  const base64Jwt = token.split(".");
  if (base64Jwt.length !== 3) return {};

  const base64Payload = base64Jwt[1];

  const decodedPayload = base64Payload.replace(/-/g, "+").replace(/_/g, "/");

  const jsonPayload = decodeURIComponent(
    atob(decodedPayload)
      .split("")
      .map(function (c) {
        return "%" + ("00" + c.charCodeAt(0).toString(16)).slice(-2);
      })
      .join(""),
  );
  return JSON.parse(jsonPayload);
}
