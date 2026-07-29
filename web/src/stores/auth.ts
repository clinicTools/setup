import { defineStore } from "pinia";
import { ref } from "vue";
import { api } from "@/lib/api";
import type { User } from "@/lib/types";

/** Hält den angemeldeten Benutzer und kapselt Login/Logout. */
export const useAuthStore = defineStore("auth", () => {
  const user = ref<User | null>(null);
  const ready = ref(false);

  async function fetchMe(): Promise<void> {
    try {
      user.value = await api.me();
    } catch {
      user.value = null;
    } finally {
      ready.value = true;
    }
  }

  async function login(username: string, password: string): Promise<void> {
    user.value = await api.login(username, password);
  }

  async function logout(): Promise<void> {
    try {
      await api.logout();
    } finally {
      user.value = null;
    }
  }

  return { user, ready, fetchMe, login, logout };
});
