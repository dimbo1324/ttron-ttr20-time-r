import "@testing-library/jest-dom";

/**
 * The browser APIs jsdom does not implement.
 *
 * Radix primitives probe pointer capture and scrolling on mount, and jsdom
 * answers with `undefined` rather than a method, which throws before a single
 * assertion runs. Stubbing them here keeps every component test free of the
 * same four lines of boilerplate.
 */
beforeAll(() => {
  Object.defineProperty(window, "matchMedia", {
    writable: true,
    value: (query: string) => ({
      matches: false,
      media: query,
      onchange: null,
      addListener: jest.fn(),
      removeListener: jest.fn(),
      addEventListener: jest.fn(),
      removeEventListener: jest.fn(),
      dispatchEvent: jest.fn(),
    }),
  });

  if (!Element.prototype.hasPointerCapture) {
    Element.prototype.hasPointerCapture = jest.fn(() => false);
    Element.prototype.setPointerCapture = jest.fn();
    Element.prototype.releasePointerCapture = jest.fn();
  }
  if (!Element.prototype.scrollIntoView) {
    Element.prototype.scrollIntoView = jest.fn();
  }

  if (!navigator.clipboard) {
    // Configurable on purpose: `userEvent.setup()` installs a clipboard stub of
    // its own, and a non-configurable one makes that throw before the test runs.
    Object.defineProperty(navigator, "clipboard", {
      writable: true,
      configurable: true,
      value: { writeText: jest.fn(async () => undefined) },
    });
  }

  global.ResizeObserver = class {
    observe() {}
    unobserve() {}
    disconnect() {}
  } as unknown as typeof ResizeObserver;
});
