// Targets the cache-event subscription helpers in common/event.js.
// The pubsub-js wrapper itself is third-party and not retested here;
// this file pins the contract specific to subscribeEntityActions —
// which actions fire, which don't, and how custom action sets and
// malformed event names behave.
import { describe, it, expect, beforeEach, vi } from "vitest";

import $event, {
  ENTITY_MUTATIONS,
  subscribeEntityActions,
  ACTION_CREATED,
  ACTION_UPDATED,
  ACTION_DELETED,
  ACTION_ARCHIVED,
  ACTION_RESTORED,
} from "common/event";

// Force pubsub-js to flush synchronously for the duration of one test.
// The module configures `async: true` for production so a handler that
// throws can't take down the publisher, but in the test the synchronous
// dispatch lets us assert call counts without flushPromises gymnastics.
async function flushEvents() {
  await new Promise((resolve) => setTimeout(resolve, 0));
}

describe("common/event.js", () => {
  describe("ENTITY_MUTATIONS", () => {
    it("contains the five verbs that match the backend constants", () => {
      // Mirrors internal/event/publish_entities.go EntityUpdated /
      // EntityCreated / EntityDeleted / EntityArchived / EntityRestored.
      // Adding a verb here MUST also land a matching constant on the
      // Go side.
      expect([...ENTITY_MUTATIONS].sort()).toEqual(
        [ACTION_ARCHIVED, ACTION_CREATED, ACTION_DELETED, ACTION_RESTORED, ACTION_UPDATED].sort()
      );
    });

    it("exports the five action verbs as individual string constants", () => {
      // Switch cases in page-level onUpdate handlers reference these
      // symbols so IDE find-references / grep locates every event-
      // handling site without false positives from prose, Vue
      // lifecycle hooks, Date methods, etc.
      expect(ACTION_CREATED).toBe("created");
      expect(ACTION_UPDATED).toBe("updated");
      expect(ACTION_DELETED).toBe("deleted");
      expect(ACTION_ARCHIVED).toBe("archived");
      expect(ACTION_RESTORED).toBe("restored");
    });

    it("is frozen so call sites cannot mutate the shared default", () => {
      expect(Object.isFrozen(ENTITY_MUTATIONS)).toBe(true);
    });
  });

  describe("subscribeEntityActions", () => {
    let token;

    beforeEach(() => {
      if (token) {
        $event.unsubscribe(token);
        token = null;
      }
    });

    it("forwards events whose action is in the default mutation set", async () => {
      const handler = vi.fn();
      token = subscribeEntityActions("test_ns_default", handler);

      for (const action of ["created", "updated", "deleted", "archived", "restored"]) {
        $event.publish(`test_ns_default.${action}`, { entities: [`uid-${action}`] });
      }
      await flushEvents();

      expect(handler).toHaveBeenCalledTimes(5);
      // Third argument is the parsed action — verifies the helper
      // extracts it correctly for downstream callers that want to
      // branch (e.g. future per-action logic).
      const actions = handler.mock.calls.map(([, , action]) => action);
      expect(actions.sort()).toEqual(["archived", "created", "deleted", "restored", "updated"]);
    });

    it("ignores unknown actions on the subscribed namespace", async () => {
      const handler = vi.fn();
      token = subscribeEntityActions("test_ns_unknown", handler);

      $event.publish("test_ns_unknown.viewed", {});
      $event.publish("test_ns_unknown.merged", { entities: ["a"] });
      $event.publish("test_ns_unknown.foo.bar", {});
      await flushEvents();

      expect(handler).not.toHaveBeenCalled();
    });

    it("ignores publishes on a sibling namespace", async () => {
      const handler = vi.fn();
      token = subscribeEntityActions("test_ns_a", handler);

      $event.publish("test_ns_b.updated", { entities: ["a"] });
      await flushEvents();

      expect(handler).not.toHaveBeenCalled();
    });

    it("respects a custom action set", async () => {
      const handler = vi.fn();
      token = subscribeEntityActions("test_ns_custom", handler, new Set(["deleted"]));

      $event.publish("test_ns_custom.updated", { entities: ["a"] });
      $event.publish("test_ns_custom.deleted", { entities: ["b"] });
      $event.publish("test_ns_custom.edited", { entities: ["c"] });
      await flushEvents();

      expect(handler).toHaveBeenCalledTimes(1);
      expect(handler).toHaveBeenCalledWith("test_ns_custom.deleted", { entities: ["b"] }, "deleted");
    });

    it("tolerates a publish on the bare namespace without an action", async () => {
      const handler = vi.fn();
      token = subscribeEntityActions("test_ns_bare", handler);

      $event.publish("test_ns_bare", { entities: ["a"] });
      await flushEvents();

      expect(handler).not.toHaveBeenCalled();
    });

    it("returns the pubsub-js token so callers can unsubscribe", async () => {
      const handler = vi.fn();
      token = subscribeEntityActions("test_ns_unsub", handler);

      $event.publish("test_ns_unsub.updated", { entities: ["a"] });
      await flushEvents();
      expect(handler).toHaveBeenCalledTimes(1);

      $event.unsubscribe(token);
      token = null;
      $event.publish("test_ns_unsub.updated", { entities: ["b"] });
      await flushEvents();
      expect(handler).toHaveBeenCalledTimes(1);
    });

    it("delivers the full topic and original payload to the handler", async () => {
      const handler = vi.fn();
      token = subscribeEntityActions("test_ns_topic", handler);

      const payload = { entities: ["one", "two"], meta: 42 };
      $event.publish("test_ns_topic.updated", payload);
      await flushEvents();

      expect(handler).toHaveBeenCalledWith("test_ns_topic.updated", payload, "updated");
    });
  });
});
