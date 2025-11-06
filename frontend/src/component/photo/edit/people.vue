<template>
  <div class="p-tab p-tab-photo-people">
    <div class="pa-2 p-faces">
      <v-alert
        v-if="markers.length === 0"
        color="surface-variant"
        icon="mdi-lightbulb-outline"
        class="no-results ma-2 opacity-70"
        variant="outlined"
      >
        <div class="font-weight-bold">
          {{ $gettext(`No people found`) }}
        </div>
        <div class="mt-2">
          {{ $gettext(`You may rescan your library to find additional faces.`) }}
          {{ $gettext(`Recognition starts after indexing has been completed.`) }}
        </div>
      </v-alert>
      <div v-else class="v-row search-results face-results cards-view d-flex">
        <div v-for="m in markers" :key="m.UID" class="v-col-12 v-col-sm-6 v-col-md-4 v-col-lg-3 d-flex">
          <v-card :data-id="m.UID" :class="m.classes()" class="result not-selectable flex-grow-1" tabindex="1">
            <v-img :src="m.thumbnailUrl('tile_320')" aspect-ratio="1" class="card">
              <v-btn
                v-if="!m.SubjUID && !m.Invalid"
                :ripple="false"
                class="input-reject"
                icon
                variant="text"
                density="comfortable"
                position="absolute"
                :title="$gettext('Remove')"
                @click.stop.prevent="onReject(m)"
              >
                <v-icon class="action-reject">mdi-close</v-icon>
              </v-btn>
              <div v-else-if="hasFaceMenu(m)" class="face-actions" data-testid="face-actions">
                <p-action-menu
                  :items="() => getFaceActions(m)"
                  button-class="input-menu"
                  list-class="opacity-85"
                ></p-action-menu>
              </div>
            </v-img>
            <v-card-actions class="meta pa-0">
              <v-btn
                v-if="m.Invalid"
                :disabled="busy"
                size="large"
                variant="flat"
                block
                :rounded="false"
                class="action-undo text-center"
                :title="$gettext('Undo')"
                @click.stop="onApprove(m)"
              >
                <v-icon>mdi-undo</v-icon>
              </v-btn>
              <v-text-field
                v-else-if="m.SubjUID"
                v-model="m.Name"
                :rules="[textRule]"
                :disabled="busy"
                :readonly="true"
                autocomplete="off"
                autocorrect="off"
                hide-details
                single-line
                clearable
                persistent-clear
                clear-icon="mdi-eject"
                density="comfortable"
                class="input-name pa-0 ma-0"
                @click:clear="onClearSubject(m)"
              ></v-text-field>
              <v-combobox
                v-else
                v-model:search="m.Name"
                :items="$config.values.people"
                item-title="Name"
                item-value="Name"
                :disabled="busy"
                return-object
                hide-no-data
                :menu-props="menuProps"
                hide-details
                single-line
                open-on-clear
                append-icon=""
                prepend-inner-icon="mdi-account-plus"
                density="comfortable"
                class="input-name pa-0 ma-0"
                @blur="onSetName(m)"
                @update:model-value="(person) => onSetPerson(m, person)"
                @keyup.enter.native="onSetName(m)"
              >
              </v-combobox>
            </v-card-actions>
          </v-card>
        </div>
      </div>
    </div>
    <p-confirm-dialog
      :visible="confirm.visible"
      icon="mdi-account-plus"
      :icon-size="42"
      :text="confirm?.model?.Name ? $gettext('Add %{s}?', { s: confirm.model.Name }) : $gettext('Add person?')"
      @close="onCancelSetName"
      @confirm="onConfirmSetName"
    ></p-confirm-dialog>
  </div>
</template>

<script>
import Marker from "model/marker";
import Subject from "model/subject";
import PConfirmDialog from "component/confirm/dialog.vue";
import PActionMenu from "component/action/menu.vue";

export default {
  name: "PTabPhotoPeople",
  components: { PConfirmDialog, PActionMenu },
  props: {
    uid: {
      type: String,
      default: "",
    },
  },
  data() {
    const view = this.$view.getData();
    return {
      view,
      markers: view.model.getMarkers(true),
      busy: false,
      disabled: !this.$config.feature("edit"),
      config: this.$config.values,
      readonly: this.$config.get("readonly"),
      confirm: {
        visible: false,
        model: new Marker(),
        text: this.$gettext("Add person?"),
      },
      menuProps: {
        closeOnClick: false,
        closeOnContentClick: true,
        openOnClick: false,
        density: "compact",
        maxHeight: 300,
      },
      textRule: (v) => {
        if (!v || !v.length) {
          return this.$gettext("Name");
        }

        return v.length <= this.$config.get("clip") || this.$gettext("Name too long");
      },
    };
  },
  watch: {
    uid: function () {
      this.refresh();
    },
  },
  methods: {
    refresh() {
      if (this.view.model) {
        this.markers = this.view.model.getMarkers(true);
      }
    },
    onReject(model) {
      if (this.busy || !model) return;

      this.busy = true;
      this.$notify.blockUI("busy");

      model.reject().finally(() => {
        this.$notify.unblockUI();
        this.busy = false;
      });
    },
    findPerson(uid) {
      const people = this.$config?.values?.people;

      if (!uid || !Array.isArray(people)) {
        return null;
      }

      return people.find((person) => person.UID === uid) || null;
    },
    updatePersonList(subject) {
      if (!subject) {
        return;
      }

      const people = this.$config?.values?.people;

      if (!Array.isArray(people)) {
        return;
      }

      const data = subject.getValues();
      const index = people.findIndex((person) => person.UID === subject.UID);
      if (index >= 0) {
        people[index] = Object.assign({}, people[index], data);
      } else {
        people.push(data);
      }
    },
    hasFaceMenu(marker) {
      return this.getFaceActions(marker).some((action) => action.visible);
    },
    getFaceActions(marker) {
      const assigned = !!marker?.SubjUID;
      const invalid = !!marker?.Invalid;
      const disabled = this.busy || this.disabled;

      return [
        {
          name: "go-to-person",
          /* icon: "mdi-account-search", */
          text: this.$gettext("Browse Pictures"),
          visible: assigned && !invalid,
          disabled,
          click: () => this.onGoToPerson(marker),
        },
        {
          name: "set-person-cover",
          /* icon: "mdi-account-check", */
          text: this.$gettext("Set as Cover Image"),
          visible: assigned && !invalid && !!marker?.Thumb,
          disabled,
          click: () => this.onSetPersonCover(marker),
        },
      ];
    },
    async loadSubject(uid) {
      try {
        return await new Subject({ UID: uid }).find(uid);
      } catch (err) {
        console.error("faces: failed loading subject", err);
        return null;
      }
    },
    async onGoToPerson(marker) {
      if (!marker?.SubjUID) {
        return;
      }

      let subject = this.findPerson(marker.SubjUID);

      if (!subject) {
        subject = await this.loadSubject(marker.SubjUID);
        if (!subject) {
          this.$notify.error(this.$gettext("Person not found"));
          return;
        }
        this.updatePersonList(subject);
      } else {
        subject = new Subject(subject);
      }

      const route = subject.route("all");
      const resolved = this.$router.resolve(route);
      this.$util.openUrl(resolved.href);
    },
    async onSetPersonCover(marker) {
      if (this.busy || !marker?.SubjUID || !marker?.Thumb) {
        return;
      }

      this.busy = true;
      this.$notify.blockUI("busy");

      try {
        let subject = this.findPerson(marker.SubjUID);

        if (subject) {
          subject = new Subject(subject);
        } else {
          subject = await this.loadSubject(marker.SubjUID);
        }

        if (!subject) {
          this.$notify.error(this.$gettext("Person not found"));
          return;
        }

        const updated = await subject.setCover(marker.Thumb);
        this.updatePersonList(updated);
        this.$notify.success(this.$gettext("Person cover updated"));
      } catch (err) {
        console.error("faces: failed setting person cover", err);
        this.$notify.error(this.$gettext("Could not update person cover"));
      } finally {
        this.$notify.unblockUI();
        this.busy = false;
      }
    },
    onApprove(model) {
      if (this.busy || !model) return;

      this.busy = true;

      model.approve().finally(() => (this.busy = false));
    },
    onClearSubject(model) {
      if (this.busy || !model) return;

      this.busy = true;
      this.$notify.blockUI("busy");

      model.clearSubject(model).finally(() => {
        this.$notify.unblockUI();
        this.busy = false;
      });
    },
    onSetPerson(model, person) {
      if (typeof person === "object" && model?.UID && person?.UID && person?.Name) {
        model.Name = person.Name;
        model.SubjUID = person.UID;
        this.setName(model);
      }

      return true;
    },
    onSetName(model) {
      if (this.busy || !model) {
        return;
      }

      const name = model?.Name;

      if (!name) {
        this.onCancelSetName();
        return;
      }

      this.confirm.model = model;

      const people = this.$config.values?.people;

      if (people) {
        const found = people.find((person) => person.Name.localeCompare(name, "en", { sensitivity: "base" }) === 0);
        if (found) {
          model.Name = found.Name;
          model.SubjUID = found.UID;
          this.setName(model);
          return;
        }
      }

      model.Name = name;
      model.SubjUID = "";
      this.confirm.visible = true;
    },
    onConfirmSetName() {
      if (!this.confirm?.model?.Name) {
        return;
      }

      this.setName(this.confirm.model);
    },
    onCancelSetName() {
      this.confirm.visible = false;
    },
    setName(model) {
      if (this.busy || !model) {
        return;
      }

      this.busy = true;
      this.$notify.blockUI("busy");

      return model.setName().finally(() => {
        this.$notify.unblockUI();
        this.busy = false;
        this.confirm.model = null;
        this.confirm.visible = false;
      });
    },
  },
};
</script>
