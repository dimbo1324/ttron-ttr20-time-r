import { readStoredSource, useSourceStore } from "./source-store";

/**
 * The source preference.
 *
 * Small, but worth its own tests for one reason: it is read from storage,
 * which can throw outright in a browser configured to block site data, and a
 * console that cannot start because a preference could not be read would be a
 * spectacularly bad trade.
 */

const STORAGE_KEY = "ft12.source";

beforeEach(() => {
  window.localStorage.clear();
  useSourceStore.setState({ source: "bench", hydrated: false });
});

describe("reading the stored preference", () => {
  it("returns the bench when nothing has been chosen", () => {
    expect(readStoredSource()).toBe("bench");
  });

  it("returns the live source when it was chosen", () => {
    window.localStorage.setItem(STORAGE_KEY, "live");

    expect(readStoredSource()).toBe("live");
  });

  it("ignores a value it does not recognise", () => {
    window.localStorage.setItem(STORAGE_KEY, "mainframe");

    expect(readStoredSource()).toBe("bench");
  });

  it("comes up on the bench when storage itself throws", () => {
    jest.spyOn(Storage.prototype, "getItem").mockImplementation(() => {
      throw new DOMException("The operation is insecure.");
    });

    expect(readStoredSource()).toBe("bench");
  });
});

describe("hydration", () => {
  it("starts on the bench so the server and the first client render agree", () => {
    // The server cannot know what this browser chose. Rendering the stored
    // source immediately would be a hydration mismatch on every reload for
    // anyone who picked live.
    expect(useSourceStore.getState().source).toBe("bench");
    expect(useSourceStore.getState().hydrated).toBe(false);
  });

  it("adopts the stored source once", () => {
    window.localStorage.setItem(STORAGE_KEY, "live");

    useSourceStore.getState().hydrate();

    expect(useSourceStore.getState().source).toBe("live");
    expect(useSourceStore.getState().hydrated).toBe(true);
  });

  it("does not undo a choice made before it ran", () => {
    window.localStorage.setItem(STORAGE_KEY, "live");
    useSourceStore.getState().hydrate();

    useSourceStore.getState().setSource("bench");
    useSourceStore.getState().hydrate();

    expect(useSourceStore.getState().source).toBe("bench");
  });
});

describe("choosing", () => {
  it("remembers the choice", () => {
    useSourceStore.getState().setSource("live");

    expect(window.localStorage.getItem(STORAGE_KEY)).toBe("live");
    expect(useSourceStore.getState().source).toBe("live");
  });

  it("ignores a choice that changes nothing", () => {
    const setItem = jest.spyOn(Storage.prototype, "setItem");

    useSourceStore.getState().setSource("bench");

    expect(setItem).not.toHaveBeenCalled();
  });

  it("still switches when the preference cannot be saved", () => {
    jest.spyOn(Storage.prototype, "setItem").mockImplementation(() => {
      throw new DOMException("QuotaExceededError");
    });

    useSourceStore.getState().setSource("live");

    // A preference that cannot be saved is not worth failing a click over.
    expect(useSourceStore.getState().source).toBe("live");
  });
});
