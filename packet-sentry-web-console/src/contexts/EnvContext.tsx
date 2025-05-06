import React, { createContext, useContext } from "react";
import type { RuntimeEnv } from "@/types/env";

const EnvContext = createContext<RuntimeEnv | undefined>(undefined);

export const EnvProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const config = window.__ENV__;

  if (!config) {
    throw new Error(
      "Runtime environment config is missing from window.__ENV__",
    );
  }

  return <EnvContext.Provider value={config}>{children}</EnvContext.Provider>;
};

export const useEnv = (): RuntimeEnv => {
  const context = useContext(EnvContext);
  if (!context) {
    throw new Error("useEnv must be used within an EnvProvider");
  }
  return context;
};
