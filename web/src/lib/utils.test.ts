import { cn } from "./utils";

describe("cn", () => {
  it("joins class names", () => {
    expect(cn("px-2", "py-1")).toBe("px-2 py-1");
  });

  it("resolves Tailwind conflicts last-wins", () => {
    expect(cn("px-2 py-1", "px-4")).toBe("py-1 px-4");
  });

  it("drops falsy values", () => {
    expect(cn("px-2", false && "hidden", undefined, null, "py-1")).toBe("px-2 py-1");
  });

  it("accepts conditional objects and arrays", () => {
    expect(cn(["px-2", { "py-1": true, hidden: false }])).toBe("px-2 py-1");
  });

  it("returns an empty string for no input", () => {
    expect(cn()).toBe("");
  });
});
