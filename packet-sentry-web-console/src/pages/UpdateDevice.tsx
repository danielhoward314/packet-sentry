import MainContentCardLayout from "@/layouts/MainContentCardLayout";
import { getDevice, updateDevice } from "@/lib/api";
import {
  GetDeviceResponse,
  InterfaceCaptureMap,
  UpdateDeviceRequest,
} from "@/types/api";
import { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { AlertCircle, Loader2 } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { useWebSocket } from "@/hooks/useWebSocket";
import { toast } from "sonner";

export default function UpdateDevicePage() {
  const { id } = useParams();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [existingDevice, setExistingDevice] =
    useState<GetDeviceResponse | null>(null);
  const [editingInterface, setEditingInterface] = useState<string | null>(null);
  const [bpfInput, setBpfInput] = useState<string>("");
  const [pendingChanges, setPendingChanges] = useState<Record<string, string>>(
    {},
  );
  useWebSocket();

  useEffect(() => {
    if (!id) return;
    const fetchData = async () => {
      try {
        const responseData = await getDevice(id);
        setExistingDevice(responseData);
      } catch (err) {
        console.error(err);
        setError("failed to fetch existing device");
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [id]);

  const handleEditClick = (iface: string) => {
    const currentBpf =
      Object.values(
        existingDevice?.interfaceBpfAssociations?.[iface]?.captures ?? {},
      )[0]?.bpf || "";
    setBpfInput(currentBpf);
    setEditingInterface(iface);
  };

  const handleSave = () => {
    if (!editingInterface || !existingDevice) return;

    setPendingChanges((prev) => ({
      ...prev,
      [editingInterface]: bpfInput.trim(),
    }));
    setEditingInterface(null);
    setBpfInput("");
  };

  const handleFinalSubmit = async () => {
    if (!id) return;

    const updatedInterfaces = new Set(Object.keys(pendingChanges));

    const mergedAssociations: Record<string, InterfaceCaptureMap> = {
      // 1. Include user-modified interfaces
      ...Object.fromEntries(
        Object.entries(pendingChanges).map(([deviceName, bpf]) => [
          deviceName,
          {
            captures: {
              [bpf]: {
                deviceName,
                bpf,
              },
            },
          },
        ]),
      ),

      // 2. Include unmodified interfaces from existingDevice, remapping capture keys to bpf values
      ...Object.fromEntries(
        Object.entries(existingDevice?.interfaceBpfAssociations || {})
          .filter(([iface]) => !updatedInterfaces.has(iface))
          .map(([iface, ifaceMap]) => [
            iface,
            {
              captures: Object.fromEntries(
                Object.values(ifaceMap.captures).map((entry) => [
                  entry.bpf,
                  entry,
                ]),
              ),
            },
          ]),
      ),
    };

    const updateRequest = {
      pcapVersion: existingDevice?.pcapVersion,
      interfaces: existingDevice?.interfaces,
      clientCertPem: existingDevice?.clientCertPem,
      clientCertFingerprint: existingDevice?.clientCertFingerprint,
      interfaceBpfAssociations: mergedAssociations,
    } as UpdateDeviceRequest;

    try {
      const response = await updateDevice(id, updateRequest);
      if (response.status < 200 || response.status >= 400) {
        toast.error(`Failed to make updates to device with id: ${id}.`);
        return;
      }
      setLoading(true);
      const responseData = await getDevice(id);
      setExistingDevice(responseData);
    } catch (e) {
      console.error(e);
      setError("failed to update device");
    } finally {
      setPendingChanges({});
      setEditingInterface(null);
      setBpfInput("");
      setLoading(false);
    }
  };

  if (error) {
    return (
      <MainContentCardLayout
        cardDescription="Make changes to the network capture configuration of a device."
        cardTitle="Update Device"
      >
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>
            Failed to load device in your organization.
          </AlertDescription>
        </Alert>
      </MainContentCardLayout>
    );
  }

  return (
    <MainContentCardLayout
      cardDescription="Make changes to a device's packet capture configuration."
      cardTitle="Update Device"
    >
      {loading ? (
        <div className="flex justify-center items-center h-screen">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      ) : (
        <div className="space-y-4">
          <div className="mb-6">
            <dl className="grid grid-cols-1 xl:grid-cols-2 gap-x-6 gap-y-4 text-md">
              <div>
                <dt className="font-medium text-muted-foreground">Device ID</dt>
                <dd className="text-foreground break-words">
                  {existingDevice?.id}
                </dd>
              </div>
              <div>
                <dt className="font-medium text-muted-foreground">
                  OS Unique ID
                </dt>
                <dd className="text-foreground break-words">
                  {existingDevice?.osUniqueIdentifier}
                </dd>
              </div>
              <div>
                <dt className="font-medium text-muted-foreground">
                  Cert Fingerprint
                </dt>
                <dd className="text-foreground break-words">
                  {existingDevice?.clientCertFingerprint}
                </dd>
              </div>
              <div>
                <dt className="font-medium text-muted-foreground">
                  PCAP Version
                </dt>
                <dd className="text-foreground break-words">
                  {existingDevice?.pcapVersion}
                </dd>
              </div>
            </dl>
          </div>

          <div className="space-y-2">
            <h3 className="text-lg font-medium">Interfaces</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {existingDevice?.interfaces.map((iface) => (
                <div
                  key={iface}
                  className="flex items-center justify-between gap-2 p-2 "
                >
                  <Badge
                    className="text-md flex-shrink-0"
                    variant={iface === editingInterface ? "default" : "outline"}
                  >
                    {iface}
                  </Badge>
                  <Button
                    variant="secondary"
                    size="sm"
                    onClick={() => handleEditClick(iface)}
                    disabled={editingInterface === iface}
                    className="flex-shrink-0"
                  >
                    Edit
                  </Button>
                </div>
              ))}
            </div>
            {editingInterface && (
              <div className="mt-6 space-y-3 p-4 rounded-xl border bg-muted/10">
                <Label htmlFor="bpf" className="text-sm font-semibold">
                  Edit BPF for{":"}
                  <span className="font-mono">{editingInterface}</span>
                </Label>
                <Input
                  id="bpf"
                  value={bpfInput}
                  onChange={(e) => setBpfInput(e.target.value)}
                  placeholder="e.g., tcp port 443 and udp"
                />
                <div className="flex items-center justify-between">
                  <div className="text-sm text-muted-foreground">
                    {/* Mocked feedback, replace with actual parsing later */}
                    {bpfInput && (
                      <div className="p-3 mt-2 rounded-md bg-muted text-xs font-mono text-muted-foreground">
                        Parsed{": "}
                        <span className="text-foreground">{bpfInput}</span>
                        {/* Later: Show syntax highlighting of protocol/ports */}
                      </div>
                    )}
                  </div>
                  <div className="space-x-2">
                    <Button onClick={handleSave}>Save BPF</Button>
                    <Button
                      variant="outline"
                      onClick={() => {
                        setEditingInterface(null);
                        setBpfInput("");
                      }}
                    >
                      Cancel
                    </Button>
                  </div>
                </div>
              </div>
            )}
            {Object.entries(pendingChanges).length > 0 && (
              <div className="mt-4 space-y-2">
                <h3 className="font-semibold">Staged Interface Changes</h3>
                {Object.entries(pendingChanges).map(([iface, bpf]) => (
                  <div key={iface} className="border p-2 rounded-md">
                    <p className="font-mono text-sm">
                      {iface}: <span>{bpf}</span>
                    </p>
                  </div>
                ))}
                <Button className="mt-2" onClick={handleFinalSubmit}>
                  Submit All Changes
                </Button>
              </div>
            )}
          </div>
        </div>
      )}
    </MainContentCardLayout>
  );
}
