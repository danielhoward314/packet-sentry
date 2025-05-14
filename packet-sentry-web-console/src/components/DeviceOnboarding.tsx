import { useState, useEffect, useMemo } from "react";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogTrigger,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { AlertCircle, ClipboardCopyIcon, Loader2 } from "lucide-react";
import { createInstallKey, getAdministrator } from "@/lib/api";
import { useAdminUser } from "@/contexts/AdminUserContext";
import { GetAdministratorResponse } from "@/types/api";

import { InstallInstructions, OSKey } from "./InstallInstructions";

export function DeviceOnboarding() {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const { adminUser, refreshAdminUser } = useAdminUser();
  const [existingAdmin, setExistingAdmin] =
    useState<GetAdministratorResponse | null>(null);
  const [step, setStep] = useState<string>("os");
  const [os, setOs] = useState<OSKey>('Unknown');
  const [installers, setInstallers] = useState<Map<string, string>>(new Map());
  const [installKey, setInstallKey] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (!adminUser?.id) {
      refreshAdminUser();
      if (!adminUser?.id) {
        console.error("admin user not in context after refresh");
        return;
      }
    }

    const fetchAdmin = async () => {
      try {
        const responseData = await getAdministrator(adminUser.id);
        setExistingAdmin(responseData);
      } catch (err) {
        console.error(err);
        setError("failed to fetch existing admin");
      } finally {
        setLoading(false);
      }
    };

    fetchAdmin();

    const getGithubReleaseAssets = async () => {
      try {
        const response = await fetch(
          "https://api.github.com/repos/danielhoward314/packet-sentry/releases/latest",
        );
        if (response.status < 200 || response.status >= 400) {
          throw new Error("non-success status code");
        }
        const data = await response.json();
        const installerMap = new Map();

        data.assets.forEach((asset: any) => {
          if (!asset?.browser_download_url) return;
          const url = asset?.browser_download_url;
          const filename = url.split("/").pop() || "";

          let os = null;
          let arch = null;

          if (filename.endsWith(".deb")) {
            os = "Ubuntu";
          } else if (filename.endsWith(".pkg")) {
            os = "macOS";
          } else if (filename.endsWith(".msi")) {
            os = "Windows";
          }

          if (filename.includes("amd64")) {
            arch = "x64";
          } else if (filename.includes("arm64")) {
            arch = "ARM64";
          }

          if (os && arch) {
            const key = `${os} ${arch}`;
            installerMap.set(key, url);
          }
        });

        setInstallers(installerMap);
      } catch (e) {
        console.error(e);
        setError("Failed to get packet-sentry-agent release data.");
      } finally {
        setLoading(false);
      }
    };

    getGithubReleaseAssets();
  }, []);

  const sortedEntries = useMemo(() => {
    const preferredOrder = ["Windows", "macOS", "Ubuntu"];
    return [...installers.entries()].sort(([a], [b]) => {
      const getPriority = (key: string) =>
        preferredOrder.findIndex((os) => key.startsWith(os));
      return getPriority(a) - getPriority(b);
    });
  }, [installers]);

  const handleDownload = (osArch: string, url: string) => {
    window.open(url, "_blank", "noopener,noreferrer");
    setOs(osArch as OSKey);
    setStep("key");
  };

  const handleCreateInstallKey = async () => {
    try {
      const response = await createInstallKey({
        administratorEmail: existingAdmin?.email ?? '',
      });

      setInstallKey(response?.installKey ?? '');
    } catch (e) {
      setError('Failed to create install key for new device.')
    }
  };

  const copyToClipboard = async () => {
    if (installKey) {
      await navigator.clipboard.writeText(installKey);
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    }
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center h-screen">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertTitle>Error</AlertTitle>
        <AlertDescription>{error}</AlertDescription>
      </Alert>
    );
  }

  return (
    <Tabs
      value={step}
      onValueChange={setStep}
      className="w-full max-w-3xl mx-auto mt-10"
    >
      <TabsList className="w-full">
        <TabsTrigger value="os">Installer Download</TabsTrigger>
        <TabsTrigger value="key" disabled={!os}>
          Key
        </TabsTrigger>
        <TabsTrigger value="summary" disabled={!installKey}>
          Summary
        </TabsTrigger>
      </TabsList>

      <TabsContent value="os">
        <Card>
          <CardContent className="p-6 grid gap-4">
            <h2 className="text-lg font-semibold">
              Download the Packet Sentry Agent
            </h2>
            <div className="flex flex-wrap gap-6">
              {sortedEntries && sortedEntries.length > 0 ? (
                sortedEntries.map(([key, url]) => (
                  <Button
                    key={key}
                    className="min-w-48"
                    onClick={() => handleDownload(key, url)}
                  >
                    {key}
                  </Button>
                ))
              ): (
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
              )}
            </div>
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="key">
        <Card>
          <CardContent className="p-6 grid gap-4">
            <Dialog>
              <DialogTrigger asChild>
                <Button onClick={handleCreateInstallKey}>
                  Generate Install Key
                </Button>
              </DialogTrigger>
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>Install Key</DialogTitle>
                </DialogHeader>
                <div className="flex items-center max-w-lg">
                  <code className="bg-muted px-2 py-1 rounded text-sm wrap-anywhere flow-text break-words">
                    {installKey || "No key yet"}
                  </code>
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={copyToClipboard}
                    disabled={!installKey}
                  >
                    <ClipboardCopyIcon className="w-4 h-4" />
                  </Button>
                  {copied && (
                    <span className="text-sm text-green-600">Copied!</span>
                  )}
                </div>
                <Button onClick={() => setStep("summary")}>Continue</Button>
              </DialogContent>
            </Dialog>
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="summary">
        {os === 'Unknown' ? (
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertTitle>Error</AlertTitle>
        <AlertDescription>This onboarding wizard has run into an error. Restart the sequence.</AlertDescription>
      </Alert>
        ) : (
        <Card>
          <CardContent className="p-6 grid gap-4">
            <InstallInstructions os={os} />
          </CardContent>
        </Card>
        )}
      </TabsContent>
    </Tabs>
  );
}
