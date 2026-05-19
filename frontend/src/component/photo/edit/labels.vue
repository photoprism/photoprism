<template>
  <div class="p-tab p-tab-photo-labels">
    <v-form ref="form" class="p-form p-form--table p-form-photo-labels" validate-on="invalid-input" accept-charset="UTF-8" tabindex="-1" @submit.prevent>
      <div class="form-body">
        <div class="form-controls">
          <v-row dense align="start">
            <v-col cols="0" sm="2" class="form-thumb">
              <div>
                <img :alt="view?.model.Title" :src="view?.model.thumbnailUrl('tile_500')" class="clickable" @click.stop.prevent.exact="openPhoto()" />
              </div>
            </v-col>
            <v-col cols="12" sm="10" class="d-flex flex-column ga-4">
              <div
                :class="$vuetify.display.smAndDown ? 'v-table--density-compact' : 'v-table--density-comfortable'"
                class="v-table v-table--has-top v-table--hover v-data-table elevation-0 edit-table list-view"
              >
                <div class="v-table__wrapper">
                  <table>
                    <thead>
                      <tr>
                        <th class="v-data-table__td v-data-table-column--align-left v-data-table__th" colspan="1" rowspan="1">
                          <div class="v-data-table-header__content">
                            <span>{{ $gettext(`Label`) }}</span>
                          </div>
                        </th>
                        <th class="v-data-table__td v-data-table-column--align-left v-data-table__th" colspan="1" rowspan="1">
                          <div class="v-data-table-header__content">
                            <span>{{ $gettext(`Source`) }}</span>
                          </div>
                        </th>
                        <th class="v-data-table__td v-data-table-column--align-center v-data-table__th" colspan="1" rowspan="1">
                          <div class="v-data-table-header__content">
                            <span>{{ $gettext(`Confidence`) }}</span>
                          </div>
                        </th>
                        <th class="v-data-table__td v-data-table-column--align-center v-data-table__th" colspan="1" rowspan="1">
                          <div class="v-data-table-header__content">
                            <span>{{ $gettext(`Action`) }}</span>
                          </div>
                        </th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="label in view.model.Labels" :key="label.LabelID" class="label result">
                        <td class="text-start">
                          {{ label.Label.Name }}
                          <!--                  TODO: add this dialog later-->
                          <!--                  <v-dialog class="p-inline-edit" @save="renameLabel(props.item.Label)">-->
                          <!--                    {{ props.item.Label.Name }}-->
                          <!--                    <template #input>-->
                          <!--                      <v-text-field v-model="props.item.Label.Name" :rules="[nameRule]" :label="$gettext('Name')" color="surface-variant" class="input-rename background-inherit elevation-0" single-line autofocus variant="solo" hide-details></v-text-field>-->
                          <!--                    </template>-->
                          <!--                  </v-dialog>-->
                        </td>
                        <td class="text-start">
                          {{ sourceName(label.LabelSrc) }}
                        </td>
                        <td class="text-center">{{ 100 - label.Uncertainty }}%</td>
                        <td class="text-center">
                          <v-btn
                            v-if="disabled"
                            icon
                            density="comfortable"
                            variant="text"
                            :ripple="false"
                            class="action-view"
                            title="Search"
                            @click.stop.prevent="searchLabel(label.Label)"
                          >
                            <v-icon color="surface-variant">mdi-magnify</v-icon>
                          </v-btn>
                          <v-btn
                            v-else-if="(label.LabelSrc === 'manual' && label.Uncertainty < 100) || (label.LabelSrc === 'batch' && label.Uncertainty === 0)"
                            icon
                            density="comfortable"
                            variant="text"
                            :ripple="false"
                            class="action-delete"
                            title="Delete"
                            @click.stop.prevent="removeLabel(label.Label)"
                          >
                            <v-icon color="surface-variant">mdi-delete</v-icon>
                          </v-btn>
                          <v-btn
                            v-else-if="label.Uncertainty < 100"
                            icon
                            density="comfortable"
                            variant="text"
                            :ripple="false"
                            class="action-remove"
                            title="Remove"
                            @click.stop.prevent="removeLabel(label.Label)"
                          >
                            <v-icon color="surface-variant">mdi-minus</v-icon>
                          </v-btn>
                          <v-btn
                            v-else
                            icon
                            density="comfortable"
                            variant="text"
                            :ripple="false"
                            class="action-on"
                            title="Activate"
                            @click.stop.prevent="activateLabel(label.Label)"
                          >
                            <v-icon color="surface-variant">mdi-plus</v-icon>
                          </v-btn>
                        </td>
                      </tr>
                      <tr v-if="!disabled" class="label result">
                        <td class="text-start">
                          <!-- No autofocus: v-combobox auto-opens its menu on
                            focus, so autofocusing here would pop the dropdown
                            the moment the Labels tab renders (and lay the menu
                            out before the tab geometry settles, which
                            mispositions it). The user clicks the input when
                            they want to type.
                            :menu-icon="null" hides the default dropdown chevron
                            because the row's density makes the chevron sit
                            visibly below the input baseline; the auto-open
                            on focus is the discovery affordance instead.
                            v-model:menu + suppressMenuOpen reproduce the
                            chip-selector trick: clearing the model after
                            committing a selection would otherwise re-open
                            the menu via the search-changed watcher. -->
                          <v-combobox
                            ref="labelInputField"
                            v-model="newLabelModel"
                            v-model:search="newLabel"
                            v-model:menu="menuOpen"
                            :items="labelOptions"
                            item-title="Name"
                            item-value="Name"
                            return-object
                            :rules="rules.text(false, 0, LabelMaxLength.Name, $gettext('Name'))"
                            color="surface-variant"
                            autocomplete="off"
                            single-line
                            flat
                            variant="plain"
                            density="compact"
                            hide-no-data
                            append-icon=""
                            :menu-icon="null"
                            :menu-props="menuProps"
                            :list-props="{ density: 'compact' }"
                            class="input-label ma-0 pa-0"
                            @focus="loadLabelOptions"
                            @update:model-value="onLabelSelected"
                            @update:menu="onMenuUpdate"
                            @keydown.enter.stop.prevent="addLabel"
                          ></v-combobox>
                        </td>
                        <td class="text-start">
                          {{ sourceName("manual") }}
                        </td>
                        <td class="text-center">100%</td>
                        <td class="text-center">
                          <v-btn icon density="comfortable" variant="text" :ripple="false" title="Add" class="p-photo-label-add" @click.stop.prevent="addLabel">
                            <v-icon color="surface-variant">mdi-plus</v-icon>
                          </v-btn>
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
            </v-col>
          </v-row>
        </div>
      </div>
    </v-form>
  </div>
</template>

<script>
import Thumb from "model/thumb";
import { MaxLength as LabelMaxLength } from "model/label";
import { rules } from "common/form";
import typeaheadCache from "common/typeahead-cache";

export default {
  name: "PTabPhotoLabels",
  props: {
    uid: {
      type: String,
      default: "",
    },
  },
  emits: ["close"],
  data() {
    return {
      view: this.$view.getData(),
      disabled: !this.$config.feature("edit"),
      config: this.$config.values,
      readonly: this.$config.get("readonly"),
      rules,
      LabelMaxLength,
      selected: [],
      newLabel: "",
      newLabelModel: null,
      // Typeahead suggestions sourced lazily from the shared cache
      // (common/typeahead-cache.js) on first focus. Cache de-dupes
      // across the sidebar combobox and the batch-edit dialog. The
      // template binds the `labelOptions` computed below, which
      // filters out labels already on the photo so the dropdown only
      // surfaces actionable suggestions.
      cachedLabelOptions: [],
      // v-model:menu binding so the post-add reset can close the menu
      // explicitly. suppressMenuOpen is a brief debounce window during
      // which onMenuUpdate refuses to re-open after a commit — Vuetify
      // would otherwise re-open via the search-changed watcher when we
      // clear newLabel/newLabelModel synchronously.
      menuOpen: false,
      suppressMenuOpen: false,
      // location="bottom start" anchors the menu's start corner to the
      // input's bottom-start (left edge in LTR, right edge in RTL —
      // logical positioning, no manual RTL handling needed). Vuetify's
      // v-combobox default is "bottom" (centered), which combined with
      // a wider-than-input min-width pushes the dropdown leftward into
      // the photo-thumbnail column. minWidth:0 lets the menu match the
      // input's natural width instead of forcing a fixed minimum.
      menuProps: {
        location: "bottom start",
        minWidth: 0,
      },
      listColumns: [
        {
          title: this.$gettext("Label"),
          key: "",
          sortable: false,
          align: "left",
        },
        {
          title: this.$gettext("Source"),
          key: "LabelSrc",
          sortable: false,
          align: "left",
        },
        {
          title: this.$gettext("Confidence"),
          key: "Uncertainty",
          sortable: false,
          align: "center",
        },
        {
          title: this.$gettext("Action"),
          key: "",
          sortable: false,
          align: "center",
        },
      ],
    };
  },
  computed: {
    // Suggestions surfaced in the combobox dropdown — the cached label
    // list with anything already assigned to this photo filtered out,
    // then sorted alphabetically via locale-aware comparison so the
    // result reads naturally across languages (e.g. Hebrew under RTL).
    // Stays reactive so a label added via this tab disappears from
    // the dropdown the moment view.model.Labels updates.
    labelOptions() {
      const normalize = (s) => (this.$util?.normalizeTitle ? this.$util.normalizeTitle(s) : (s || "").toLowerCase());
      const assigned = Array.isArray(this.view?.model?.Labels) ? this.view.model.Labels : [];
      const assignedKeys = new Set(assigned.map((l) => normalize(l?.Label?.Name)).filter(Boolean));
      return this.cachedLabelOptions
        .filter((l) => !assignedKeys.has(normalize(l.Name)))
        .slice()
        .sort((a, b) => (a.Name || "").localeCompare(b.Name || "", undefined, { sensitivity: "base", numeric: true }));
    },
  },
  methods: {
    refresh() {},
    sourceName(s) {
      return this.$util.sourceName(s);
    },
    removeLabel(label) {
      if (!label || !this.view?.model) {
        return;
      }

      const name = label.Name;

      this.view.model.removeLabel(label.ID).then(() => {
        this.$notify.success("removed " + name);
      });
    },
    addLabel() {
      const typed = (this.newLabel || "").trim();
      if (!typed || !this.view?.model) {
        return;
      }

      // Apply the same canonical-match dedup the sidebar uses for L3:
      // typing `Hello Cat` resolves to an existing `hello-cat` label so
      // the backend isn't asked to create a near-duplicate. normalizeTitle
      // ignores case and converts every punctuation character to
      // whitespace.
      const normalize = (s) => (this.$util.normalizeTitle ? this.$util.normalizeTitle(s) : (s || "").toLowerCase());
      const norm = normalize(typed);
      let finalName = typed;
      if (norm) {
        const existing = this.labelOptions.find((l) => normalize(l.Name) === norm);
        if (existing) {
          finalName = existing.Name;
        }
      }

      // Already on the photo? Skip the API call so picking an existing
      // chip from the dropdown doesn't bounce through addLabel and
      // surface a stray "Label updated" + "added <name>" notification
      // pair (the backend treats a re-add as an update). Match against
      // the photo's own Labels using the same normalization so typed
      // variants like `EARTH` collapse onto an existing `Earth`.
      if (norm) {
        const labels = Array.isArray(this.view.model.Labels) ? this.view.model.Labels : [];
        const alreadyOnPhoto = labels.some((l) => normalize(l?.Label?.Name) === norm);
        if (alreadyOnPhoto) {
          this.resetInput();
          return;
        }
      }

      this.view.model.addLabel(finalName).then(() => {
        this.$notify.success("added " + finalName);
        this.resetInput();
      });
    },
    // Selecting an existing label from the dropdown commits via the same
    // canonical-name path as addLabel — keeps the chip name consistent
    // with what's stored server-side.
    onLabelSelected(value) {
      if (!value || typeof value !== "object" || !value.Name) {
        return;
      }
      this.newLabel = value.Name;
      this.addLabel();
    },
    // Pulls suggestions from the shared cache on first focus. Cheap on
    // re-focus (cache hit) and refreshes after WS-driven evictions.
    // Stored on the raw `cachedLabelOptions` ref; the `labelOptions`
    // computed filters out anything already assigned to the photo.
    loadLabelOptions() {
      typeaheadCache
        .getLabels()
        .then((models) => {
          this.cachedLabelOptions = models.map((l) => ({ Name: l.Name, UID: l.UID }));
        })
        .catch(() => {});
    },
    // Closes the dropdown, blurs the input, then clears the bound
    // values. Mirrors the chip-selector pattern: clearing inside an
    // open combobox would otherwise re-open the menu via the
    // search-changed watcher (the "" search is treated as "show all"
    // and pops the dropdown again).
    resetInput() {
      this.menuOpen = false;
      this.suppressMenuOpen = true;
      this.$nextTick(() => {
        const input = this.$refs.labelInputField;
        if (input && typeof input.blur === "function") {
          input.blur();
        }
        this.newLabel = "";
        this.newLabelModel = null;
        window.setTimeout(() => {
          this.suppressMenuOpen = false;
        }, 200);
      });
    },
    // Vetoes Vuetify's "open the menu" intent during the post-commit
    // debounce window so a stale focus event can't re-pop the dropdown
    // immediately after the user picked an item.
    onMenuUpdate(val) {
      if (val && this.suppressMenuOpen) {
        this.menuOpen = false;
        return;
      }
      this.menuOpen = val;
    },
    activateLabel(label) {
      if (!label || !this.view?.model) {
        return;
      }

      this.view.model.activateLabel(label.ID);
    },
    // TODO: add this dialog later
    // renameLabel(label) {
    //   if (!label) {
    //     return;
    //   }
    //
    //   this.view.model.renameLabel(label.ID, label.Name);
    // },
    searchLabel(label) {
      this.$router.push({ name: "all", query: { q: "label:" + label.Slug } }).catch(() => {});
      this.$emit("close");
    },
    openPhoto() {
      if (!this.view?.model) {
        return;
      }

      this.$lightbox.openModels(Thumb.fromPhotos([this.view.model]), 0);
    },
  },
};
</script>
