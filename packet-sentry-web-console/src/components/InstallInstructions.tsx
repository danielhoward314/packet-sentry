import { INSTALLER_INSTRUCTIONS } from "@/lib/consts";

export type OSKey = keyof typeof INSTALLER_INSTRUCTIONS;

interface InstallInstructionsProps {
  os: OSKey;
}

export function InstallInstructions({ os }: InstallInstructionsProps) {
  return (
    <>
      <h2 className="text-lg font-semibold">Summary</h2>
      <p>
        You've downloaded the installer for {os}. Use the instructions below to
        run it on the device. When prompted, enter the install key.
      </p>
      {os === "Windows x64" ? (
        <pre className="bg-muted p-2 rounded text-sm whitespace-pre-wrap">
          {INSTALLER_INSTRUCTIONS[os]}
        </pre>
      ) : (
        <pre className="bg-muted p-2 rounded text-sm whitespace-pre-wrap">
        <code className="bg-muted px-2 py-1 rounded text-sm">
          {INSTALLER_INSTRUCTIONS[os]}
        </code>
        </pre>
      )}
    </>
  );
}
