import MainContentCardLayout from "@/layouts/MainContentCardLayout";
import { getDevice } from "@/lib/api";
import { GetDeviceResponse } from "@/types/api";
import { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { AlertCircle, Loader2 } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

export default function UpdateDevicePage() {
  const { id } = useParams();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [existingDevice, setExistingDevice] = useState<GetDeviceResponse | null>(null);

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
      cardDescription="Make changes to an administrator in your organization."
      cardTitle="Update Administrator"
    >
      {loading ? (
        <div className="flex justify-center items-center h-screen">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      ) : (
        <div>
        <p>{existingDevice?.id}</p>
        <p>{existingDevice?.osUniqueIdentifier}</p>
        <p>{existingDevice?.clientCertFingerprint}</p>
        <p>{existingDevice?.clientCertPem}</p>
        <p>{existingDevice?.interfaces}</p>
        <p>{existingDevice?.pcapVersion}</p>
        </div>
      )}
    </MainContentCardLayout>
  );
}
