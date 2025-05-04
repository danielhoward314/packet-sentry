export async function validateJwt(token: string): Promise<boolean> {
  // Simulate network delay
  await new Promise(resolve => setTimeout(resolve, 500))

  // Fake logic: only this token is valid
  return token === 'valid-token'
}
