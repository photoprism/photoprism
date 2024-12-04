<template>
  <div id="p-navigation" :class="{ 'sidenav-visible': drawer }">
    <template v-if="visible && $vuetify.display.smAndDown">
      <v-toolbar position="fixed" flat dense color="navigation darken-1" class="nav-small elevation-2" @click.stop.prevent>
        <v-avatar class="bg-transparent navigation-logo clickable" tile :size="28" :class="{ clickable: auth }" @click.stop.prevent="showNavigation()">
          <img :src="appIcon" :alt="appName" :class="{ 'animate-hue': indexing }" />
        </v-avatar>
        <v-toolbar-title class="nav-title">
          <span :class="{ clickable: auth }" @click.stop.prevent="showNavigation()">{{ page.title }}</span>
        </v-toolbar-title>
        <v-btn :ripple="false" class="mobile-menu-trigger elevation-0 rounded-circle" @click.stop.prevent="speedDial = true">
          <v-icon>mdi-dots-vertical</v-icon>
        </v-btn>
      </v-toolbar>
    </template>
    <template v-else-if="visible && !auth">
      <v-toolbar flat dense color="navigation darken-1" class="nav-small">
        <v-avatar class="bg-transparent navigation-logo" tile :size="28">
          <img :src="appIcon" :alt="appName" />
        </v-avatar>
        <v-toolbar-title class="nav-title">
          {{ page.title }}
        </v-toolbar-title>
        <v-btn :ripple="false" class="mobile-menu-trigger elevation-0 rounded-circle" @click.stop.prevent="speedDial = true">
          <v-icon>mdi-dots-vertical</v-icon>
        </v-btn>
      </v-toolbar>
    </template>
    <v-navigation-drawer v-if="visible && auth" v-model="drawer" color="navigation" :rail="isMini" :width="270" :mobile-breakpoint="960" :rail-width="80" class="nav-sidebar navigation p-flex-nav" :location="rtl ? 'right' : undefined">
      <v-toolbar flat :dense="$vuetify.display.smAndDown">
        <v-list class="navigation-home elevation-0" bg-color="navigation-home" width="100%">
          <v-list-item class="px-3" :elevation="0" :ripple="false" @click.stop.prevent="goHome">
            <template #prepend>
              <div class="v-avatar bg-transparent navigation-logo clickable" @click.stop.prevent="goHome">
                <img :src="appIcon" :alt="appName" :class="{ 'animate-hue': indexing }" />
              </div>
            </template>
            <template #append>
              <v-btn icon variant="text" :elevation="0" class="nav-minimize hidden-sm-and-down" :ripple="false" :title="$gettext('Minimize')" @click.stop="toggleIsMini()">
                <v-icon v-if="!rtl">mdi-chevron-left</v-icon>
                <v-icon v-else>mdi-chevron-right</v-icon>
              </v-btn>
            </template>
            <v-list-item-title class="nav-appname mr-auto">
              {{ appName }}
            </v-list-item-title>
          </v-list-item>
        </v-list>
      </v-toolbar>

      <v-list nav class="pt-3 p-flex-menu navigation-menu" bg-color="navigation" color="primary" open-strategy="single">
        <v-list-item v-if="isMini && !isRestricted" class="nav-expand" @click.stop="toggleIsMini()">
          <v-icon v-if="!rtl" class="ma-auto">mdi-chevron-right</v-icon>
          <v-icon v-else class="ma-auto">mdi-chevron-left</v-icon>
        </v-list-item>

        <v-list-item v-if="isMini && $config.feature('search')" to="/browse" variant="text" class="nav-browse" :ripple="false" @click.stop="">
          <v-icon class="ma-auto">mdi-magnify</v-icon>
        </v-list-item>
        <div v-else-if="!isMini && $config.feature('search')">
          <v-list-item to="/browse" variant="text" class="nav-browse activator" :ripple="false" @click.stop="">
            <v-list-item-title class="p-flex-menuitem">
              <p class="nav-item-title">
                <translate key="Search">Search</translate>
              </p>
              <!-- TODO: fix filter -->
              <!-- <span v-if="config.count.all > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.all | abbreviateCount }}</span> -->
              <span v-if="config.count.all > 0" :class="`nav-count-group ${rtl ? '--rtl' : ''}`">{{ config.count.all }}</span>
            </v-list-item-title>
          </v-list-item>

          <v-list-group>
            <template #activator="{ props }">
              <v-list-item v-bind="props" variant="text" class="nav-browse activator-parent" :ripple="false" @click.stop="">
                <v-icon>mdi-magnify</v-icon>
              </v-list-item>
            </template>

            <v-list-item :to="{ name: 'browse', query: { q: 'mono:true quality:3 photo:true' } }" :exact="true" variant="text" class="nav-monochrome" :ripple="false" @click.stop="">
              <v-list-item-title :class="`p-flex-menuitem menu-item ${rtl ? '--rtl' : ''}`">
                <translate>Monochrome</translate>
              </v-list-item-title>
            </v-list-item>

            <v-list-item :to="{ name: 'browse', query: { q: 'panoramas' } }" :exact="true" variant="text" class="nav-panoramas" :ripple="false" @click.stop="">
              <v-list-item-title :class="`p-flex-menuitem menu-item ${rtl ? '--rtl' : ''}`">
                <translate>Panoramas</translate>
              </v-list-item-title>
            </v-list-item>

            <v-list-item :to="{ name: 'browse', query: { q: 'animated' } }" :exact="true" variant="text" class="nav-animated" :ripple="false" @click.stop="">
              <v-list-item-title :class="`p-flex-menuitem menu-item ${rtl ? '--rtl' : ''}`">
                <translate>Animated</translate>
              </v-list-item-title>
            </v-list-item>

            <v-list-item v-show="isSponsor" :to="{ name: 'browse', query: { q: 'vectors' } }" :exact="true" variant="text" class="nav-vectors" :ripple="false" @click.stop="">
              <v-list-item-title :class="`p-flex-menuitem menu-item ${rtl ? '--rtl' : ''}`">
                <translate>Vectors</translate>
              </v-list-item-title>
            </v-list-item>

            <v-list-item :to="{ name: 'photos', query: { q: 'stacks' } }" :exact="true" variant="text" class="nav-stacks" :ripple="false" @click.stop="">
              <v-list-item-title :class="`p-flex-menuitem menu-item ${rtl ? '--rtl' : ''}`">
                <translate>Stacks</translate>
              </v-list-item-title>
            </v-list-item>

            <v-list-item :to="{ name: 'photos', query: { q: 'scans' } }" :exact="true" variant="text" class="nav-scans" :ripple="false" @click.stop="">
              <v-list-item-title :class="`p-flex-menuitem menu-item ${rtl ? '--rtl' : ''}`">
                <translate>Scans</translate>
              </v-list-item-title>
            </v-list-item>

            <v-list-item v-if="canManagePhotos" v-show="$config.feature('review')" to="/review" variant="text" class="nav-review" :ripple="false" @click.stop="">
              <v-list-item-title :class="`p-flex-menuitem menu-item ${rtl ? '--rtl' : ''}`">
                <translate>Review</translate>
              </v-list-item-title>
              <!-- TODO: fix filter -->
              <span v-show="config.count.review > 0" :class="`nav-count-item ${rtl ? '--rtl' : ''}`">{{ config.count.review }}</span>
              <!-- <span v-show="config.count.review > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.review | abbreviateCount }}</span> -->
            </v-list-item>

            <v-list-item v-show="$config.feature('archive')" to="/archive" variant="text" class="nav-archive" :ripple="false" @click.stop="">
              <v-list-item-title :class="`p-flex-menuitem menu-item ${rtl ? '--rtl' : ''}`">
                <translate>Archive</translate>
              </v-list-item-title>
              <!-- TODO: fix filter -->
              <span v-show="config.count.archived > 0" :class="`nav-count-item ${rtl ? '--rtl' : ''}`">{{ config.count.archived }}</span>
              <!-- <span v-show="config.count.archived > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.archived | abbreviateCount }}</span> -->
            </v-list-item>
          </v-list-group>
        </div>

        <v-list-item v-if="isMini" v-show="$config.feature('albums')" to="/albums" variant="text" class="nav-albums" :ripple="false" @click.stop="">
          <v-icon class="ma-auto">mdi-bookmark</v-icon>
        </v-list-item>
        <div v-else-if="!isMini" v-show="$config.feature('albums')">
          <v-list-item to="/albums" variant="text" class="nav-albums activator" :ripple="false" @click.stop="">
            <v-list-item-title class="p-flex-menuitem">
              <p class="nav-item-title">
                <translate key="Albums">Albums</translate>
              </p>
              <!-- TODO: fix filter -->
              <!-- <span v-if="config.count.albums > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.albums | abbreviateCount }}</span> -->
              <span v-if="config.count.albums > 0" :class="`nav-count-group ${rtl ? '--rtl' : ''}`">{{ config.count.albums }}</span>
            </v-list-item-title>
          </v-list-item>

          <v-list-group>
            <template #activator="{ props }">
              <v-list-item v-bind="props" to="{ name: 'albums' }" variant="text" class="nav-albums activator-parent" :ripple="false" @click.stop="">
                <v-icon>mdi-bookmark</v-icon>
              </v-list-item>
            </template>

            <v-list-item to="/unsorted" variant="text" class="nav-unsorted" :ripple="false">
              <v-list-item-title :class="`p-flex-menuitem menu-item ${rtl ? '--rtl' : ''}`">
                <translate key="Unsorted">Unsorted</translate>
              </v-list-item-title>
            </v-list-item>
          </v-list-group>
        </div>

        <v-list-item v-if="isMini && $config.feature('videos')" to="/videos" variant="text" class="nav-video" :ripple="false" @click.stop="">
          <v-icon class="ma-auto">mdi-play-circle</v-icon>
        </v-list-item>
        <div v-else-if="!isMini && $config.feature('videos')">
          <v-list-item to="/videos" variant="text" class="nav-video activator" :ripple="false" @click.stop="">
            <v-list-item-title class="p-flex-menuitem">
              <p class="nav-item-title">
                <translate key="Videos">Videos</translate>
              </p>
              <!-- TODO: fix filter -->
              <!-- <span v-show="config.count.videos > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.videos | abbreviateCount }}</span> -->
              <span v-show="config.count.videos > 0" :class="`nav-count-group ${rtl ? '--rtl' : ''}`">{{ config.count.videos }}</span>
            </v-list-item-title>
          </v-list-item>

          <v-list-group>
            <template #activator="{ props }">
              <v-list-item v-bind="props" to="/videos" variant="text" class="nav-video" :ripple="false" @click.stop="">
                <v-icon>mdi-play-circle</v-icon>
              </v-list-item>
            </template>

            <v-list-item :to="{ name: 'live' }" variant="text" class="nav-live" :ripple="false" @click.stop="">
              <v-list-item-title :class="`p-flex-menuitem menu-item ${rtl ? '--rtl' : ''}`">
                <translate>Live Photos</translate>
              </v-list-item-title>
              <!-- TODO: fix filter -->
              <span v-show="config.count.live > 0" :class="`nav-count-item ${rtl ? '--rtl' : ''}`">{{ config.count.live }}</span>
              <!-- <span v-show="config.count.live > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.live | abbreviateCount }}</span> -->
            </v-list-item>
          </v-list-group>
        </div>

        <v-list-item v-if="isMini && $config.feature('people') && (canManagePeople || config.count.people > 0)" :to="{ name: 'people' }" variant="text" class="nav-people" :ripple="false" @click.stop="">
          <v-icon class="ma-auto">mdi-account</v-icon>
        </v-list-item>
        <v-list-item v-else-if="!isMini && $config.feature('people') && (canManagePeople || config.count.people > 0)" :to="{ name: 'people' }" variant="text" class="nav-people" :ripple="false" @click.stop="">
          <v-list-item-title class="p-flex-menuitem">
            <v-icon>mdi-account</v-icon>
            <p class="nav-item-title">
              <translate key="People">People</translate>
            </p>
          </v-list-item-title>
          <!-- TODO: fix filter -->
          <span v-show="config.count.people > 0" :class="`nav-count-item ${rtl ? '--rtl' : ''}`">{{ config.count.people }}</span>
          <!-- <span v-show="config.count.people > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.people | abbreviateCount }}</span> -->
        </v-list-item>

        <v-list-item v-if="isMini && $config.feature('favorites')" :to="{ name: 'favorites' }" variant="text" class="nav-favorites" :ripple="false" @click.stop="">
          <v-icon class="ma-auto">mdi-heart</v-icon>
        </v-list-item>
        <v-list-item v-else-if="!isMini && $config.feature('favorites')" :to="{ name: 'favorites' }" variant="text" class="nav-favorites" :ripple="false" @click.stop="">
          <v-list-item-title class="p-flex-menuitem">
            <v-icon>mdi-heart</v-icon>
            <p class="nav-item-title">
              <translate key="Favorites">Favorites</translate>
            </p>
          </v-list-item-title>
          <!-- TODO: fix filter -->
          <span v-show="config.count.favorites > 0" :class="`nav-count-item ${rtl ? '--rtl' : ''}`">{{ config.count.favorites }}</span>
          <!-- <span v-show="config.count.favorites > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.favorites | abbreviateCount }}</span> -->
        </v-list-item>

        <v-list-item v-if="isMini && $config.feature('moments')" :to="{ name: 'moments' }" variant="text" class="nav-moments" :ripple="false" @click.stop="">
          <v-icon class="ma-auto">mdi-star</v-icon>
        </v-list-item>
        <v-list-item v-else-if="!isMini && $config.feature('moments')" :to="{ name: 'moments' }" variant="text" class="nav-moments" :ripple="false" @click.stop="">
          <v-list-item-title class="p-flex-menuitem">
            <v-icon>mdi-star</v-icon>
            <p class="nav-item-title">
              <translate key="Moments">Moments</translate>
            </p>
          </v-list-item-title>
          <!-- TODO: fix filter -->
          <span v-show="config.count.moments > 0" :class="`nav-count-item ${rtl ? '--rtl' : ''}`">{{ config.count.moments }}</span>
          <!-- <span v-show="config.count.moments > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.moments | abbreviateCount }}</span> -->
        </v-list-item>

        <v-list-item v-if="isMini && $config.feature('moments')" :to="{ name: 'calendar' }" variant="text" class="nav-calendar" :ripple="false" @click.stop="">
          <v-icon class="ma-auto">mdi-calendar-range</v-icon>
        </v-list-item>
        <v-list-item v-else-if="!isMini && $config.feature('moments')" :to="{ name: 'calendar' }" variant="text" class="nav-calendar" :ripple="false" @click.stop="">
          <v-list-item-title class="p-flex-menuitem">
            <v-icon>mdi-calendar-range</v-icon>
            <p class="nav-item-title">
              <translate key="Calendar">Calendar</translate>
            </p>
          </v-list-item-title>
          <!-- TODO: fix filter -->
          <span v-show="config.count.months > 0" :class="`nav-count-item ${rtl ? '--rtl' : ''}`">{{ config.count.months }}</span>
          <!-- <span v-show="config.count.months > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.months | abbreviateCount }}</span> -->
        </v-list-item>

        <v-list-item v-if="isRestricted" v-show="$config.feature('places')" to="/states" variant="text" class="nav-states" :ripple="false" @click.stop="">
          <v-list-item-title class="p-flex-menuitem" @click.stop="">
            <!-- TODO: change the icon -->
            <v-icon>near_me</v-icon>
            <translate key="States">States</translate>
          </v-list-item-title>
          <!-- TODO: fix filter -->
          <span v-show="config.count.states > 0" :class="`nav-count-item ${rtl ? '--rtl' : ''}`">{{ config.count.states }}</span>
          <!-- <span v-show="config.count.states > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.states | abbreviateCount }}</span> -->
        </v-list-item>

        <template v-if="canSearchPlaces">
          <v-list-item v-if="isMini" v-show="canSearchPlaces && $config.feature('places')" :to="{ name: 'places' }" variant="text" class="nav-places" :ripple="false" @click.stop="">
            <v-icon class="ma-auto">mdi-map-marker</v-icon>
          </v-list-item>
          <div v-else v-show="canSearchPlaces && $config.feature('places')">
            <v-list-item to="/places" variant="text" class="nav-places activator" :ripple="false" @click.stop="">
              <v-list-item-title class="p-flex-menuitem">
                <p class="nav-item-title">
                  <translate key="Places">Places</translate>
                </p>
                <!-- TODO: fix filter -->
                <span v-show="config.count.places > 0" :class="`nav-count-group ${rtl ? '--rtl' : ''}`">{{ config.count.places }}</span>
                <!-- <span v-show="config.count.places > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.places | abbreviateCount }}</span> -->
              </v-list-item-title>
            </v-list-item>

            <v-list-group>
              <template #activator="{ props }">
                <v-list-item v-bind="props" to="/places" variant="text" class="nav-places" :ripple="false" @click.stop="">
                  <v-icon>mdi-map-marker</v-icon>
                </v-list-item>
              </template>

              <v-list-item to="/states" variant="text" class="nav-states" @click.stop="">
                <v-list-item-title :class="`p-flex-menuitem menu-item ${rtl ? '--rtl' : ''}`">
                  <translate key="States">States</translate>
                </v-list-item-title>
                <!-- TODO: fix filter -->
                <span v-show="config.count.states > 0" :class="`nav-count-item ${rtl ? '--rtl' : ''}`">{{ config.count.states }}</span>
                <!-- <span v-show="config.count.states > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.states | abbreviateCount }}</span> -->
              </v-list-item>
            </v-list-group>
          </div>
        </template>

        <v-list-item v-if="isMini && $config.feature('labels')" to="/labels" variant="text" class="nav-labels" :ripple="false" @click.stop="">
          <v-icon class="ma-auto">mdi-label</v-icon>
        </v-list-item>
        <v-list-item v-else-if="!isMini && $config.feature('labels')" to="/labels" variant="text" class="nav-labels" :ripple="false" @click.stop="">
          <v-list-item-title class="p-flex-menuitem">
            <v-icon>mdi-label</v-icon>
            <p class="nav-item-title">
              <translate key="Labels">Labels</translate>
            </p>
          </v-list-item-title>
          <!-- TODO: fix filter -->
          <span v-show="config.count.labels > 0" :class="`nav-count-item ${rtl ? '--rtl' : ''}`">{{ config.count.labels }}</span>
          <!-- <span v-show="config.count.labels > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.labels | abbreviateCount }}</span> -->
        </v-list-item>

        <v-list-item v-if="isMini && $config.feature('folders')" :to="{ name: 'folders' }" variant="text" class="nav-folders" :ripple="false" @click.stop="">
          <v-icon class="ma-auto">mdi-folder</v-icon>
        </v-list-item>
        <v-list-item v-else-if="!isMini && $config.feature('folders')" :to="{ name: 'folders' }" variant="text" class="nav-folders" :ripple="false" @click.stop="">
          <v-list-item-title class="p-flex-menuitem">
            <v-icon>mdi-folder</v-icon>
            <p class="nav-item-title">
              <translate key="Folders">Folders</translate>
            </p>
          </v-list-item-title>
          <!-- TODO: fix filter -->
          <span v-show="config.count.folders > 0" :class="`nav-count-item ${rtl ? '--rtl' : ''}`">{{ config.count.folders }}</span>
          <!-- <span v-show="config.count.folders > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.folders | abbreviateCount }}</span> -->
        </v-list-item>

        <v-list-item v-if="isMini && $config.feature('private')" to="/private" variant="text" class="nav-private" :ripple="false" @click.stop="">
          <v-icon class="ma-auto">mdi-lock</v-icon>
        </v-list-item>
        <v-list-item v-else-if="!isMini && $config.feature('private')" to="/private" variant="text" class="nav-private" :ripple="false" @click.stop="">
          <v-list-item-title class="p-flex-menuitem">
            <v-icon>mdi-lock</v-icon>
            <p class="nav-item-title">
              <translate key="Private">Private</translate>
            </p>
          </v-list-item-title>
          <!-- TODO: fix filter -->
          <span v-show="config.count.private > 0" :class="`nav-count-item ${rtl ? '--rtl' : ''}`">{{ config.count.private }}</span>
          <!-- <span v-show="config.count.private > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.private | abbreviateCount }}</span> -->
        </v-list-item>

        <v-list-item v-if="isMini && $config.feature('library')" :to="{ name: 'library_index' }" variant="text" class="nav-library" :ripple="false" @click.stop="">
          <v-icon class="ma-auto">mdi-film</v-icon>
        </v-list-item>
        <div v-else-if="!isMini && $config.feature('library')">
          <v-list-item :to="{ name: 'library_index' }" variant="text" class="nav-library activator" :ripple="false" @click.stop="">
            <v-list-item-title class="p-flex-menuitem">
              <p class="nav-item-title">
                <translate key="Library">Library</translate>
              </p>
            </v-list-item-title>
          </v-list-item>

          <v-list-group>
            <template #activator="{ props }">
              <v-list-item v-bind="props" :to="{ name: 'library_index' }" variant="text" class="nav-library" :ripple="false" @click.stop="">
                <v-icon>mdi-film</v-icon>
              </v-list-item>
            </template>

            <v-list-item v-show="$config.feature('files')" to="/index/files" variant="text" class="nav-originals" :ripple="false" @click.stop="">
              <v-list-item-title :class="`p-flex-menuitem menu-item ${rtl ? '--rtl' : ''}`">
                <translate key="Originals">Originals</translate>
              </v-list-item-title>
              <!-- TODO: fix filter -->
              <span v-show="config.count.files > 0 && canAccessPrivate" :class="`nav-count-item ${rtl ? '--rtl' : ''}`">{{ config.count.files }}</span>
              <!-- <span v-show="config.count.files > 0 && canAccessPrivate" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.files | abbreviateCount }}</span> -->
            </v-list-item>

            <v-list-item :to="{ name: 'hidden' }" variant="text" class="nav-hidden" :ripple="false" @click.stop="">
              <v-list-item-title :class="`p-flex-menuitem menu-item ${rtl ? '--rtl' : ''}`">
                <translate key="Hidden">Hidden</translate>
              </v-list-item-title>
              <!-- TODO: fix filter -->
              <span v-show="config.count.hidden > 0" :class="`nav-count-item ${rtl ? '--rtl' : ''}`">{{ config.count.hidden }}</span>
              <!-- <span v-show="config.count.hidden > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.hidden | abbreviateCount }}</span> -->
            </v-list-item>

            <v-list-item :to="{ name: 'errors' }" variant="text" class="nav-errors" @click.stop="">
              <v-list-item-title :class="`p-flex-menuitem menu-item ${rtl ? '--rtl' : ''}`">
                <translate key="Errors">Errors</translate>
              </v-list-item-title>
            </v-list-item>
          </v-list-group>
        </div>

        <template v-if="!config.disable.settings">
          <v-list-item v-if="isMini" v-show="$config.feature('settings')" :to="{ name: 'settings' }" variant="text" class="nav-settings" :ripple="false" @click.stop="">
            <v-icon class="ma-auto">mdi-cog</v-icon>
          </v-list-item>
          <div v-else-if="!isMini" v-show="$config.feature('settings')">
            <v-list-item :to="{ name: 'settings' }" variant="text" class="nav-settings activator" :ripple="false" @click.stop="">
              <v-list-item-title class="p-flex-menuitem">
                <p class="nav-item-title">
                  <translate key="Settings">Settings</translate>
                </p>
              </v-list-item-title>
            </v-list-item>

            <v-list-group>
              <template #activator="{ props }">
                <v-list-item v-bind="props" :to="{ name: 'settings' }" variant="text" class="nav-settings" :ripple="false" @click.stop="">
                  <v-icon>mdi-cog</v-icon>
                </v-list-item>
              </template>

              <v-list-item v-if="canManageUsers" :to="{ path: '/admin/users' }" :exact="false" variant="text" class="nav-admin-users" :ripple="false" @click.stop="">
                <v-list-item-title :class="`menu-item ${rtl ? '--rtl' : ''}`">
                  <translate>Users</translate>
                </v-list-item-title>
              </v-list-item>

              <v-list-item v-show="featFeedback" :to="{ name: 'feedback' }" :exact="true" variant="text" class="nav-feedback" :ripple="false" @click.stop="">
                <v-list-item-title :class="`menu-item ${rtl ? '--rtl' : ''}`">
                  <translate>Feedback</translate>
                </v-list-item-title>
              </v-list-item>

              <v-list-item :to="{ name: 'license' }" :exact="true" variant="text" class="nav-license" :ripple="false" @click.stop="">
                <v-list-item-title :class="`menu-item ${rtl ? '--rtl' : ''}`">
                  <translate key="License">License</translate>
                </v-list-item-title>
              </v-list-item>

              <v-list-item v-show="featUpgrade" :to="{ name: 'upgrade' }" variant="text" class="nav-upgrade" :exact="true" :ripple="false" @click.stop="">
                <v-list-item-title :class="`menu-item ${rtl ? '--rtl' : ''}`">
                  <translate key="Upgrade">Upgrade</translate>
                </v-list-item-title>
              </v-list-item>

              <v-list-item :to="{ name: 'about' }" :exact="true" variant="text" class="nav-about" :ripple="false" @click.stop="">
                <v-list-item-title :class="`menu-item ${rtl ? '--rtl' : ''}`">
                  <translate>About</translate>
                </v-list-item-title>
              </v-list-item>
            </v-list-group>
          </div>
        </template>

        <v-list-item v-show="!auth" :to="{ name: 'login' }" variant="text" class="nav-login" @click.stop="">
          <v-list-item-title>
            <v-icon>mdi-lock</v-icon>
            <translate key="Login">Login</translate>
          </v-list-item-title>
        </v-list-item>

        <v-list-item v-if="featMembership" :to="{ name: 'upgrade' }" variant="text" class="nav-upgrade" @click.stop="">
          <v-list-item-title v-if="isPro" class="p-flex-menuitem">
            <v-icon>mdi-check-circle</v-icon>
            <p class="nav-item-title">
              <translate key="Upgrade">Upgrade</translate>
            </p>
          </v-list-item-title>
          <v-list-item-title v-else class="p-flex-menuitem">
            <v-icon>mdi-diamond</v-icon>
            <p class="nav-item-title">
              <translate key="Support Our Mission">Support Our Mission</translate>
            </p>
          </v-list-item-title>
        </v-list-item>
      </v-list>

      <v-container v-if="$config.disconnected" class="bg-navigation position-fixed bottom-0 border-t-thin pa-0">
        <v-row no-gutters class="nav-connecting ma-0 py-2 clickable" align="center" @click.stop="showServerConnectionHelp">
          <v-col :cols="isMini ? 12 : 3">
            <div class="text-center">
              <v-icon color="warning" size="28">mdi-wifi-off</v-icon>
            </div>
          </v-col>
          <v-col v-if="!isMini" cols="9">
            <p class="text-left text-body-2">
              <translate key="No server connection">No server connection</translate>
            </p>
          </v-col>
        </v-row>
      </v-container>
      <v-container v-else-if="auth && !isPublic" class="p-user-box bg-navigation p-profile position-fixed bottom-0 border-t-thin">
        <v-row no-gutters align="center">
          <v-col :cols="isMini ? 12 : 3">
            <div class="navigation-user-avatar clickable text-center my-2" @click.stop="showAccountSettings">
              <img :src="userAvatarURL" :alt="accountInfo" :title="accountInfo" class="rounded-circle" />
            </div>
          </v-col>
          <v-col v-if="!isMini" cols="6">
            <div class="text-left mt-1">
              <p class="text-body-2">{{ displayName }}</p>
              <p class="text-caption opacity-70">{{ accountInfo }}</p>
            </div>
          </v-col>
          <v-col :cols="isMini ? 12 : 3">
            <div class="text-center my-1">
              <v-btn icon variant="text" :elevation="0" @click.stop.prevent="onLogout">
                <v-icon>mdi-power</v-icon>
              </v-btn>
            </div>
          </v-col>
        </v-row>
      </v-container>
    </v-navigation-drawer>
    <div id="mobile-menu" :class="{ active: speedDial }" @click.stop="speedDial = false">
      <div class="menu-content grow-top-right">
        <div class="menu-icons">
          <a v-if="auth && !isPublic" href="#" :title="$gettext('Logout')" class="menu-action navigation-logout" @click.prevent="onLogout">
            <v-icon>mdi-power</v-icon>
          </a>
          <a href="#" :title="$gettext('Reload')" class="menu-action nav-reload" @click.prevent="reloadApp">
            <v-icon>mdi-refresh</v-icon>
          </a>
          <router-link v-if="auth && $config.feature('account')" :to="{ name: 'settings_account' }" :title="$gettext('Account')" class="menu-action nav-account">
            <v-icon>mdi-shield-account-variant</v-icon>
          </router-link>
          <router-link v-if="auth && $config.feature('settings') && !routeName('settings')" :to="{ name: 'settings' }" :title="$gettext('Settings')" class="menu-action nav-settings">
            <v-icon>mdi-cog</v-icon>
          </router-link>
          <a v-if="auth && !config.readonly && $config.feature('upload')" href="#" :title="$gettext('Upload')" class="menu-action nav-upload" @click.prevent="openUpload()">
            <v-icon>mdi-cloud-upload</v-icon>
          </a>
          <router-link v-if="!auth && !isPublic" :to="{ name: 'login' }" :title="$gettext('Login')" class="menu-action nav-login">
            <!-- TODO: change this icon -->
            <v-icon>login</v-icon>
          </router-link>
        </div>
        <div class="menu-actions">
          <div v-if="auth && !routeName('browse') && $config.feature('search')" class="menu-action nav-search">
            <router-link to="/browse">
              <v-icon>mdi-magnify</v-icon>
              <translate>Search</translate>
            </router-link>
          </div>
          <div v-if="auth && !routeName('albums') && $config.feature('albums')" class="menu-action nav-albums">
            <router-link to="/albums">
              <v-icon>mdi-bookmark</v-icon>
              <translate>Albums</translate>
            </router-link>
          </div>
          <div v-if="auth && canManagePeople && !routeName('people') && $config.feature('people')" class="menu-action nav-people">
            <router-link to="/people">
              <v-icon>mdi-account</v-icon>
              <translate>People</translate>
            </router-link>
          </div>
          <div v-if="auth && canSearchPlaces && !routeName('places') && $config.feature('places')" class="menu-action nav-places">
            <router-link to="/places">
              <v-icon>mdi-map-marker</v-icon>
              <translate>Places</translate>
            </router-link>
          </div>
          <div v-if="auth && !routeName('files') && $config.feature('files') && $config.feature('library')" class="menu-action nav-files">
            <router-link to="/index/files">
              <v-icon>mdi-folder</v-icon>
              <translate>Files</translate>
            </router-link>
          </div>
          <div v-if="auth && !routeName('library_index') && $config.feature('library')" class="menu-action nav-index">
            <router-link :to="{ name: 'library_index' }">
              <v-icon>mdi-film</v-icon>
              <translate>Index</translate>
            </router-link>
          </div>
          <div v-if="auth && !routeName('index') && $config.feature('library') && $config.feature('logs')" class="menu-action nav-logs">
            <router-link :to="{ name: 'library_logs' }">
              <v-icon>mdi-file-document</v-icon>
              <translate>Logs</translate>
            </router-link>
          </div>
          <div class="menu-action nav-manual">
            <a href="https://link.photoprism.app/docs" target="_blank">
              <v-icon>mdi-book-open-page-variant</v-icon>
              <translate>User Guide</translate>
            </a>
          </div>
          <div v-if="featUpgrade" class="menu-action nav-upgrade">
            <router-link :to="{ name: 'upgrade' }">
              <template #default="{ href, navigate, isActive }">
                <a :href="href" :class="{ active: isActive }" @click="navigate">
                  <v-icon v-if="isPro">mdi-check-circle</v-icon>
                  <v-icon v-else>mdi-diamond</v-icon>
                  <translate>Upgrade</translate>
                </a>
              </template>
            </router-link>
          </div>
          <div v-if="config.legalUrl" class="menu-action nav-legal">
            <a :href="config.legalUrl" target="_blank">
              <v-icon>mdi-information</v-icon>
              <translate>Legal Information</translate>
            </a>
          </div>
        </div>
      </div>
    </div>
    <div v-if="config.legalInfo && visible" id="legal-info">
      <span v-if="config.legalUrl" class="clickable" @click.stop.prevent="showLegalInfo()">{{ config.legalInfo }}</span>
      <span v-else>{{ config.legalInfo }}</span>
    </div>
    <p-reload-dialog :show="reload.dialog" @close="reload.dialog = false"></p-reload-dialog>
    <p-upload-dialog :show="upload.dialog" :data="upload.data" @cancel="upload.dialog = false" @confirm="upload.dialog = false"></p-upload-dialog>
    <p-photo-edit-dialog :show="edit.dialog" :selection="edit.selection" :index="edit.index" :album="edit.album" @close="edit.dialog = false"></p-photo-edit-dialog>
  </div>
</template>

<script>
import Event from "pubsub-js";
import Album from "model/album";

export default {
  name: "PNavigation",
  filters: {
    abbreviateCount: (val) => {
      const value = Number.parseInt(val);
      // TODO: make abbreviation configurable by userprofile settings or env var before enabling it.
      // if (value >= 1000) {
      //   const digits = value % 1000 <= 50 ? 0 : 1;
      //   return (value/1000).toFixed(digits).toString()+'k';
      // }
      return value;
    },
  },
  data() {
    const appName = this.$config.getName();

    let appNameSuffix = "";
    let appNameParts = appName.split(" ");

    if (appNameParts.length > 1) {
      appNameSuffix = appNameParts.slice(1, 9).join(" ");
    }

    const isDemo = this.$config.get("demo");
    const isPro = !!this.$config.values?.ext["pro"];
    const isPublic = this.$config.get("public");
    const isReadOnly = this.$config.get("readonly");
    const isRestricted = this.$config.deny("photos", "access_library");
    const isSuperAdmin = this.$session.isSuperAdmin();
    const tier = this.$config.getTier();

    return {
      canSearchPlaces: this.$config.allow("places", "search"),
      canAccessPrivate: !isRestricted && this.$config.allow("photos", "access_private"),
      canManagePhotos: this.$config.allow("photos", "manage"),
      canManagePeople: this.$config.allow("people", "manage"),
      canManageUsers: (!isPublic || isDemo) && this.$config.allow("users", "manage"),
      appNameSuffix: appNameSuffix,
      appName: this.$config.getName(),
      appAbout: this.$config.getAbout(),
      appIcon: this.$config.getIcon(),
      indexing: false,
      drawer: null,
      featUpgrade: tier < 6 && isSuperAdmin && !isPublic && !isDemo,
      featMembership: tier < 3 && isSuperAdmin && !isPublic && !isDemo,
      featFeedback: tier >= 6 && isSuperAdmin && !isPublic && !isDemo,
      isRestricted: isRestricted,
      isMini: localStorage.getItem("last_navigation_mode") !== "false" || isRestricted,
      isDemo: isDemo,
      isPro: isPro,
      isPublic: isPublic,
      isReadOnly: isReadOnly,
      isAdmin: this.$session.isAdmin(),
      isSuperAdmin: isSuperAdmin,
      isSponsor: this.$config.isSponsor(),
      isTest: this.$config.test,
      session: this.$session,
      config: this.$config.values,
      page: this.$config.page,
      user: this.$session.getUser(),
      reload: {
        dialog: false,
      },
      upload: {
        dialog: false,
        data: {},
      },
      edit: {
        dialog: false,
        album: null,
        selection: [],
        index: 0,
      },
      speedDial: false,
      rtl: this.$rtl,
      subscriptions: [],
    };
  },
  computed: {
    auth() {
      return this.session.auth || this.isPublic;
    },
    visible() {
      return !this.$route.meta.hideNav;
    },
    displayName() {
      const user = this.$session.getUser();
      if (user) {
        return user.getDisplayName();
      }

      return this.$gettext("Unregistered");
    },
    userAvatarURL() {
      return this.$session.getUser().getAvatarURL("tile_50");
    },
    accountInfo() {
      const user = this.$session.getUser();
      if (user) {
        return user.getAccountInfo();
      }

      return this.$gettext("Account");
    },
  },
  created() {
    this.subscriptions.push(Event.subscribe("index", this.onIndex));
    this.subscriptions.push(Event.subscribe("import", this.onIndex));
    this.subscriptions.push(Event.subscribe("dialog.reload", () => (this.reload.dialog = true)));
    this.subscriptions.push(
      Event.subscribe("dialog.upload", (ev, data) => {
        if (data) {
          this.upload.data = data;
        } else {
          this.upload.data = {};
        }
        this.upload.dialog = true;
      })
    );
    this.subscriptions.push(
      Event.subscribe("dialog.edit", (ev, data) => {
        if (!this.edit.dialog) {
          this.edit.dialog = true;
          this.edit.index = data.index;
          this.edit.selection = data.selection;
          this.edit.album = data.album;
        }
      })
    );
  },
  unmounted() {
    for (let i = 0; i < this.subscriptions.length; i++) {
      Event.unsubscribe(this.subscriptions[i]);
    }
  },
  methods: {
    routeName(name) {
      if (!name || !this.$route.name) {
        return false;
      }

      return this.$route.name.startsWith(name);
    },
    reloadApp() {
      this.$notify.info(this.$gettext("Reloading…"));
      this.$notify.blockUI();
      setTimeout(() => window.location.reload(), 100);
    },
    openUpload() {
      if (this.auth && !this.isReadOnly && this.$config.feature("upload")) {
        if (this.$route.name === "album" && this.$route.params?.album) {
          return new Album()
            .find(this.$route.params?.album)
            .then((m) => {
              this.upload.dialog = true;
              this.upload.data = { albums: [m] };
            })
            .catch(() => {
              this.upload.dialog = true;
              this.upload.data = { albums: [] };
            });
        } else {
          this.upload.dialog = true;
          this.upload.data = { albums: [] };
        }
      } else {
        this.goHome();
      }
    },
    goHome() {
      if (this.$route.name !== "home") {
        this.$router.push({ name: "home" });
      }
    },
    showNavigation() {
      if (this.auth) {
        this.drawer = true;
        this.isMini = this.isRestricted;
      }
    },
    toggleIsMini() {
      this.isMini = !this.isMini;
      localStorage.setItem("last_navigation_mode", `${this.isMini}`);
    },
    showAccountSettings: function () {
      if (this.$config.feature("account")) {
        this.$router.push({ name: "settings_account" });
      } else {
        this.$router.push({ name: "settings" });
      }
    },
    showServerConnectionHelp: function () {
      this.$router.push({ path: "/help/websockets" });
    },
    showLegalInfo() {
      if (this.config.legalUrl) {
        window.open(this.config.legalUrl, "_blank").focus();
      } else {
        this.$router.push({ name: "about" });
      }
    },
    onLogout() {
      this.$session.logout();
    },
    onIndex(ev) {
      if (!ev) {
        return;
      }

      const type = ev.split(".")[1];

      switch (type) {
        case "file":
        case "folder":
        case "indexing":
          this.indexing = true;
          break;
        case "completed":
          this.indexing = false;
          break;
      }
    },
  },
};
</script>
