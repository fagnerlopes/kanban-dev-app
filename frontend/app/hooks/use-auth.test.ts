import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { useAuth } from "./use-auth";

describe("useAuth", () => {
  beforeEach(() => {
    localStorage.clear();
    vi.restoreAllMocks();
  });

  it("starts logged out with an empty store", () => {
    const { result } = renderHook(() => useAuth());
    expect(result.current.isAuthenticated).toBe(false);
    expect(result.current.session).toBeNull();
  });

  // The regression: login used to POST /api/dev/login, a route that only
  // exists when DEV_MODE=1 — so the demo login 404'd in every deployed
  // environment.
  it("logs in without calling the network", async () => {
    const fetchSpy = vi.spyOn(globalThis, "fetch");
    const { result } = renderHook(() => useAuth());

    await act(async () => {
      await result.current.login();
    });

    expect(fetchSpy).not.toHaveBeenCalled();
    expect(result.current.isAuthenticated).toBe(true);
    expect(result.current.session?.user).toBe("demo@kanban.local");
  });

  it("restores the session from localStorage on a fresh mount", async () => {
    const first = renderHook(() => useAuth());
    await act(async () => {
      await first.result.current.login();
    });

    const second = renderHook(() => useAuth());
    expect(second.result.current.isAuthenticated).toBe(true);
    expect(second.result.current.session?.token).toBe(
      first.result.current.session?.token,
    );
  });

  it("clears the session on logout", async () => {
    const { result } = renderHook(() => useAuth());
    await act(async () => {
      await result.current.login();
    });

    act(() => {
      result.current.logout();
    });

    expect(result.current.isAuthenticated).toBe(false);
    expect(localStorage.getItem("kbd-token")).toBeNull();
  });

  it("still signs in when localStorage is unavailable", async () => {
    vi.spyOn(Storage.prototype, "setItem").mockImplementation(() => {
      throw new Error("blocked");
    });
    const { result } = renderHook(() => useAuth());

    await act(async () => {
      await result.current.login();
    });

    expect(result.current.isAuthenticated).toBe(true);
  });
});
