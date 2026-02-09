"use client"

import { createContext, useContext } from "react"
import { SessionProvider as NextAuthSessionProvider, useSession as useNextAuthSession } from "next-auth/react"

const disableAuth = process.env.NEXT_PUBLIC_DISABLE_AUTH !== "false"

const noAuthValue = { data: null, status: "unauthenticated" as const }
const NoAuthContext = createContext(noAuthValue)

export function SessionProvider({ children }: { children: React.ReactNode }) {
  if (disableAuth) {
    return <NoAuthContext.Provider value={noAuthValue}>{children}</NoAuthContext.Provider>
  }
  return <NextAuthSessionProvider>{children}</NextAuthSessionProvider>
}

function useNoAuthSession() {
  return useContext(NoAuthContext)
}

// Drop-in replacement for next-auth/react useSession.
// disableAuth is a build-time constant so the unused branch is dead-code eliminated.
export const useSession = disableAuth ? useNoAuthSession : useNextAuthSession
