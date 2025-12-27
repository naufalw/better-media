import { createContext, useContext } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "./api";

export interface User {
  id: string;
  email: string;
  name: string;
  role: string;
  created_at: string;
}

interface AuthContextType {
  user: User | null;
  isLoading: boolean;
  login: (token: string, user: User) => void;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const queryClient = useQueryClient();

  const { data: user, isLoading } = useQuery<User | null>({
    queryKey: ["auth", "me"],
    queryFn: async () => {
      const token = localStorage.getItem("auth_token");
      if (!token) return null;
      try {
        return await api.getMe();
      } catch {
        localStorage.removeItem("auth_token");
        return null;
      }
    },
    retry: false,
    staleTime: Infinity,
  });

  const login = (token: string, userData: User) => {
    localStorage.setItem("auth_token", token);
    queryClient.setQueryData(["auth", "me"], userData);
  };

  const logout = () => {
    localStorage.removeItem("auth_token");
    queryClient.setQueryData(["auth", "me"], null);
  };

  return (
    <AuthContext.Provider value={{ user: user || null, isLoading, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
