<template>
  <div id="p-navigation" :class="{ 'sidenav-visible': drawer }">
    <template v-if="visible && $vuetify.display.smAndDown">
      <v-toolbar theme="dark" position="fixed" flat dense color="navigation darken-1" class="nav-small elevation-2" @click.stop.prevent>
        <v-avatar class="nav-avatar" tile :size="28" :class="{ clickable: auth }" @click.stop.prevent="showNavigation()">
          <img :src="appIcon" :alt="config.name" :class="{ 'animate-hue': indexing }" />
        </v-avatar>
        <v-toolbar-title class="nav-title">
          <span :class="{ clickable: auth }" @click.stop.prevent="showNavigation()">{{ page.title }}</span>
        </v-toolbar-title>
        <v-btn theme="dark" :ripple="false" class="mobile-menu-trigger elevation-0 rounded-circle" @click.stop.prevent="speedDial = true">
          <v-icon>mdi-dots-vertical</v-icon>
        </v-btn>
      </v-toolbar>
    </template>
    <template v-else-if="visible && !auth">
      <v-toolbar theme="dark" flat dense color="navigation darken-1" class="nav-small">
        <v-avatar class="nav-avatar" tile :size="28">
          <img :src="appIcon" :alt="config.name" />
        </v-avatar>
        <v-toolbar-title class="nav-title">
          {{ page.title }}
        </v-toolbar-title>
        <v-btn theme="dark" :ripple="false" class="mobile-menu-trigger elevation-0 rounded-circle" @click.stop.prevent="speedDial = true">
          <v-icon>mdi-dots-vertical</v-icon>
        </v-btn>
      </v-toolbar>
    </template>
    <v-navigation-drawer v-if="visible && auth" v-model="drawer" :rail="isMini" :width="270" :mobile-breakpoint="960" :rail-width="80" class="nav-sidebar navigation p-flex-nav" :location="rtl ? 'right' : undefined">
      <v-toolbar flat :dense="$vuetify.display.smAndDown">
        <v-list class="navigation-home">
          <v-list-item class="nav-logo">
            <div class="d-flex align-center w-100">
              <img class="nav-avatar clickable mr-3" heigth="40px" width="40px" :src="appIcon" :alt="appName" :class="{ 'animate-hue': indexing }" @click.stop.prevent="goHome" />
              <v-list-item-title class="text-h6 mr-auto">
                {{ appName }}
              </v-list-item-title>
              <v-list-item-action class="hidden-sm-and-down ml-2" :title="$gettext('Minimize')">
                <v-btn icon class="nav-minimize" @click.stop="toggleIsMini()">
                  <v-icon v-if="!rtl">mdi-chevron-left</v-icon>
                  <v-icon v-else>mdi-chevron-right</v-icon>
                </v-btn>
              </v-list-item-action>
            </div>
          </v-list-item>
        </v-list>
      </v-toolbar>

      <v-list class="pt-6 p-flex-menu navigation-menu">
        <v-list-item v-if="isMini && !isRestricted" class="nav-expand" @click.stop="toggleIsMini()">
          <v-icon v-if="!rtl" class="ma-auto">mdi-chevron-right</v-icon>
          <v-icon v-else class="ma-auto">mdi-chevron-left</v-icon>
        </v-list-item>

        <v-list-item v-if="isMini && $config.feature('search')" to="/browse" class="nav-browse" @click.stop="">
          <v-icon class="ma-auto">mdi-magnify</v-icon>
        </v-list-item>
        <v-list-group v-else-if="!isMini && $config.feature('search')">
          <template #activator="{ props }">
            <v-list-item v-bind="props" to="/browse" class="nav-browse" @click.stop="">
              <v-list-item-title class="p-flex-menuitem">
                <v-icon>mdi-magnify</v-icon>
                <translate key="Search" class="nav-item-title">Search</translate>
              </v-list-item-title>
              <!-- TODO: fix filter -->
              <span v-if="config.count.all > 0" :class="`nav-count-group ${rtl ? '--rtl' : ''}`">{{ config.count.all }}</span>
              <!-- <span v-if="config.count.all > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.all | abbreviateCount }}</span> -->
            </v-list-item>
          </template>

          <v-list-item :to="{ name: 'browse', query: { q: 'mono:true quality:3 photo:true' } }" :exact="true" class="nav-monochrome" @click.stop="">
            <v-list-item-title :class="`p-flex-menuitem menu-item ${rtl ? '--rtl' : ''}`">
              <translate>Monochrome</translate>
            </v-list-item-title>
          </v-list-item>

          <v-list-item :to="{ name: 'browse', query: { q: 'panoramas' } }" :exact="true" class="nav-panoramas" @click.stop="">
            <v-list-item-title :class="`menu-item ${rtl ? '--rtl' : ''}`">
              <translate>Panoramas</translate>
            </v-list-item-title>
          </v-list-item>

          <v-list-item :to="{ name: 'browse', query: { q: 'animated' } }" :exact="true" class="nav-animated" @click.stop="">
            <v-list-item-title :class="`menu-item ${rtl ? '--rtl' : ''}`">
              <translate>Animated</translate>
            </v-list-item-title>
          </v-list-item>

          <v-list-item v-show="isSponsor" :to="{ name: 'browse', query: { q: 'vectors' } }" :exact="true" class="nav-vectors" @click.stop="">
            <v-list-item-title :class="`menu-item ${rtl ? '--rtl' : ''}`">
              <translate>Vectors</translate>
            </v-list-item-title>
          </v-list-item>

          <v-list-item :to="{ name: 'photos', query: { q: 'stacks' } }" :exact="true" class="nav-stacks" @click.stop="">
            <v-list-item-title :class="`p-flex-menuitem menu-item ${rtl ? '--rtl' : ''}`">
              <translate>Stacks</translate>
            </v-list-item-title>
          </v-list-item>

          <v-list-item :to="{ name: 'photos', query: { q: 'scans' } }" :exact="true" class="nav-scans" @click.stop="">
            <v-list-item-title :class="`p-flex-menuitem menu-item ${rtl ? '--rtl' : ''}`">
              <translate>Scans</translate>
            </v-list-item-title>
          </v-list-item>

          <v-list-item v-if="canManagePhotos" v-show="$config.feature('review')" to="/review" class="nav-review" @click.stop="">
            <v-list-item-title :class="`p-flex-menuitem menu-item ${rtl ? '--rtl' : ''}`">
              <translate>Review</translate>
            </v-list-item-title>
            <!-- TODO: fix filter -->
            <span v-show="config.count.review > 0" :class="`nav-count-item ${rtl ? '--rtl' : ''}`">{{ config.count.review }}</span>
            <!-- <span v-show="config.count.review > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.review | abbreviateCount }}</span> -->
          </v-list-item>

          <v-list-item v-show="$config.feature('archive')" to="/archive" class="nav-archive" @click.stop="">
            <v-list-item-title :class="`p-flex-menuitem menu-item ${rtl ? '--rtl' : ''}`">
              <translate>Archive</translate>
            </v-list-item-title>
            <!-- TODO: fix filter -->
            <span v-show="config.count.archived > 0" :class="`nav-count-item ${rtl ? '--rtl' : ''}`">{{ config.count.archived }}</span>
            <!-- <span v-show="config.count.archived > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.archived | abbreviateCount }}</span> -->
          </v-list-item>
        </v-list-group>

        <v-list-item v-if="isMini" v-show="$config.feature('albums')" to="/albums" class="nav-albums" @click.stop="">
          <v-icon class="ma-auto">mdi-bookmark</v-icon>
        </v-list-item>
        <v-list-group v-else-if="!isMini" v-show="$config.feature('albums')">
          <template #activator="{ props }">
            <v-list-item v-bind="props" to="{ name: 'albums' }" class="nav-albums" @click.stop="">
              <v-list-item-title class="p-flex-menuitem">
                <v-icon>mdi-bookmark</v-icon>
                <translate key="Albums" class="nav-item-title">Albums</translate>
              </v-list-item-title>
              <!-- TODO: fix filter -->
              <span v-if="config.count.albums > 0" :class="`nav-count-group ${rtl ? '--rtl' : ''}`">{{ config.count.albums }}</span>
              <!-- <span v-if="config.count.albums > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.albums | abbreviateCount }}</span> -->
            </v-list-item>
          </template>

          <v-list-item to="/unsorted" class="nav-unsorted">
            <v-list-item-title :class="`p-flex-menuitem menu-item ${rtl ? '--rtl' : ''}`">
              <translate key="Unsorted">Unsorted</translate>
            </v-list-item-title>
          </v-list-item>
        </v-list-group>

        <v-list-item v-if="isMini && $config.feature('videos')" to="/videos" class="nav-video" @click.stop="">
            <v-icon class="ma-auto">mdi-play-circle</v-icon>
        </v-list-item>
        <v-list-group v-else-if="!isMini && $config.feature('videos')">
          <template #activator="{ props }">
            <v-list-item v-bind="props" to="/videos" class="nav-video" @click.stop="">
              <v-list-item-title class="p-flex-menuitem">
                <v-icon>mdi-play-circle</v-icon>
                <translate key="Videos" class="nav-item-title">Videos</translate>
              </v-list-item-title>
              <!-- TODO: fix filter -->
              <span v-show="config.count.videos > 0" :class="`nav-count-group ${rtl ? '--rtl' : ''}`">{{ config.count.videos }}</span>
              <!-- <span v-show="config.count.videos > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.videos | abbreviateCount }}</span> -->
            </v-list-item>
          </template>

          <v-list-item :to="{ name: 'live' }" class="nav-live" @click.stop="">
            <v-list-item-title :class="`p-flex-menuitem menu-item ${rtl ? '--rtl' : ''}`">
              <translate>Live Photos</translate>
            </v-list-item-title>
            <!-- TODO: fix filter -->
            <span v-show="config.count.live > 0" :class="`nav-count-item ${rtl ? '--rtl' : ''}`">{{ config.count.live }}</span>
            <!-- <span v-show="config.count.live > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.live | abbreviateCount }}</span> -->
          </v-list-item>
        </v-list-group>

        <v-list-item v-if="isMini && $config.feature('people') && (canManagePeople || config.count.people > 0)" :to="{ name: 'people' }" class="nav-people" @click.stop="">
          <v-icon class="ma-auto">mdi-account</v-icon>
        </v-list-item>
        <v-list-item v-else-if="!isMini && $config.feature('people') && (canManagePeople || config.count.people > 0)" :to="{ name: 'people' }" class="nav-people" @click.stop="">
          <v-list-item-title class="p-flex-menuitem">
            <v-icon>mdi-account</v-icon>
            <translate key="People" class="nav-item-title">People</translate>
          </v-list-item-title>
          <!-- TODO: fix filter -->
          <span v-show="config.count.people > 0" :class="`nav-count-item ${rtl ? '--rtl' : ''}`">{{ config.count.people }}</span>
          <!-- <span v-show="config.count.people > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.people | abbreviateCount }}</span> -->
        </v-list-item>

        <v-list-item v-if="isMini && $config.feature('favorites')" :to="{ name: 'favorites' }" class="nav-favorites" @click.stop="">
          <v-icon class="ma-auto">mdi-heart</v-icon>
        </v-list-item>
        <v-list-item v-else-if="!isMini && $config.feature('favorites')" :to="{ name: 'favorites' }" class="nav-favorites" @click.stop="">
          <v-list-item-title class="p-flex-menuitem">
            <v-icon>mdi-heart</v-icon>
            <translate key="Favorites" class="nav-item-title">Favorites</translate>
          </v-list-item-title>
          <!-- TODO: fix filter -->
          <span v-show="config.count.favorites > 0" :class="`nav-count-item ${rtl ? '--rtl' : ''}`">{{ config.count.favorites }}</span>
          <!-- <span v-show="config.count.favorites > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.favorites | abbreviateCount }}</span> -->
        </v-list-item>

        <v-list-item v-if="isMini && $config.feature('moments')" :to="{ name: 'moments' }" class="nav-moments" @click.stop="">
          <v-icon class="ma-auto">mdi-star</v-icon>
        </v-list-item>
        <v-list-item v-else-if="!isMini && $config.feature('moments')" :to="{ name: 'moments' }" class="nav-moments" @click.stop="">
          <v-list-item-title class="p-flex-menuitem">
            <v-icon>mdi-star</v-icon>
            <translate key="Moments" class="nav-item-title">Moments</translate>
          </v-list-item-title>
          <!-- TODO: fix filter -->
          <span v-show="config.count.moments > 0" :class="`nav-count-item ${rtl ? '--rtl' : ''}`">{{ config.count.moments }}</span>
          <!-- <span v-show="config.count.moments > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.moments | abbreviateCount }}</span> -->
        </v-list-item>

        <v-list-item v-if="isMini && $config.feature('moments')" :to="{ name: 'calendar' }" class="nav-calendar" @click.stop="">
          <v-icon class="ma-auto">mdi-calendar-range</v-icon>
        </v-list-item>
        <v-list-item v-else-if="!isMini && $config.feature('moments')" :to="{ name: 'calendar' }" class="nav-calendar" @click.stop="">
          <v-list-item-title class="p-flex-menuitem">
            <v-icon>mdi-calendar-range</v-icon>
            <translate key="Calendar" class="nav-item-title">Calendar</translate>
          </v-list-item-title>
          <!-- TODO: fix filter -->
          <span v-show="config.count.months > 0" :class="`nav-count-item ${rtl ? '--rtl' : ''}`">{{ config.count.months }}</span>
          <!-- <span v-show="config.count.months > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.months | abbreviateCount }}</span> -->
        </v-list-item>

        <v-list-item v-if="isRestricted" v-show="$config.feature('places')" to="/states" class="nav-states" @click.stop="">
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
          <v-list-item v-if="isMini" v-show="canSearchPlaces && $config.feature('places')" :to="{ name: 'places' }" class="nav-places" @click.stop="">
            <v-icon class="ma-auto">mdi-map-marker</v-icon>
          </v-list-item>
          <v-list-group v-else v-show="canSearchPlaces && $config.feature('places')">
            <template #activator="{ props }">
              <v-list-item v-bind="props" to="/places" class="nav-places" @click.stop="">
                <v-list-item-title class="p-flex-menuitem">
                  <v-icon>mdi-map-marker</v-icon>
                  <translate key="Places" class="nav-item-title">Places</translate>
                </v-list-item-title>
                <!-- TODO: fix filter -->
                <span v-show="config.count.places > 0" :class="`nav-count-group ${rtl ? '--rtl' : ''}`">{{ config.count.places }}</span>
                <!-- <span v-show="config.count.places > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.places | abbreviateCount }}</span> -->
              </v-list-item>
            </template>

            <v-list-item to="/states" class="nav-states" @click.stop="">
              <v-list-item-title :class="`p-flex-menuitem menu-item ${rtl ? '--rtl' : ''}`">
                <translate key="States">States</translate>
              </v-list-item-title>
              <!-- TODO: fix filter -->
              <span v-show="config.count.states > 0" :class="`nav-count-item ${rtl ? '--rtl' : ''}`">{{ config.count.states }}</span>
              <!-- <span v-show="config.count.states > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.states | abbreviateCount }}</span> -->
            </v-list-item>
          </v-list-group>
        </template>

        <v-list-item v-if="isMini && $config.feature('labels')" to="/labels" class="nav-labels" @click.stop="">
          <v-icon class="ma-auto">mdi-label</v-icon>
        </v-list-item>
        <v-list-item v-else-if="!isMini && $config.feature('labels')" to="/labels" class="nav-labels" @click.stop="">
          <v-list-item-title class="p-flex-menuitem">
            <v-icon>mdi-label</v-icon>
            <translate key="Labels" class="nav-item-title">Labels</translate>
          </v-list-item-title>
          <!-- TODO: fix filter -->
          <span v-show="config.count.labels > 0" :class="`nav-count-item ${rtl ? '--rtl' : ''}`">{{ config.count.labels }}</span>
          <!-- <span v-show="config.count.labels > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.labels | abbreviateCount }}</span> -->
        </v-list-item>

        <v-list-item v-if="isMini && $config.feature('folders')" :to="{ name: 'folders' }" class="nav-folders" @click.stop="">
          <v-icon class="ma-auto">mdi-folder</v-icon>
        </v-list-item>
        <v-list-item v-else-if="!isMini && $config.feature('folders')" :to="{ name: 'folders' }" class="nav-folders" @click.stop="">
          <v-list-item-title class="p-flex-menuitem">
            <v-icon>mdi-folder</v-icon>
            <translate key="Folders" class="nav-item-title">Folders</translate>
          </v-list-item-title>
          <!-- TODO: fix filter -->
          <span v-show="config.count.folders > 0" :class="`nav-count-item ${rtl ? '--rtl' : ''}`">{{ config.count.folders }}</span>
          <!-- <span v-show="config.count.folders > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.folders | abbreviateCount }}</span> -->
        </v-list-item>

        <v-list-item v-if="isMini && $config.feature('private')" to="/private" class="nav-private" @click.stop="">
          <v-icon class="ma-auto">mdi-lock</v-icon>
        </v-list-item>
        <v-list-item v-else-if="!isMini && $config.feature('private')" to="/private" class="nav-private" @click.stop="">
          <v-list-item-title class="p-flex-menuitem">
            <v-icon>mdi-lock</v-icon>
            <translate key="Private" class="nav-item-title">Private</translate>
          </v-list-item-title>
          <!-- TODO: fix filter -->
          <span v-show="config.count.private > 0" :class="`nav-count-item ${rtl ? '--rtl' : ''}`">{{ config.count.private }}</span>
          <!-- <span v-show="config.count.private > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.private | abbreviateCount }}</span> -->
        </v-list-item>

        <v-list-item v-if="isMini && $config.feature('library')" :to="{ name: 'library_index' }" class="nav-library" @click.stop="">
          <v-icon class="ma-auto">mdi-film</v-icon>
        </v-list-item>
        <v-list-group v-else-if="!isMini && $config.feature('library')">
          <template #activator="{ props }">
            <v-list-item v-bind="props" :to="{ name: 'library_index' }" class="nav-library" @click.stop="">
              <v-list-item-title class="p-flex-menuitem">
                <v-icon>mdi-film</v-icon>
                <translate key="Library" class="nav-item-title">Library</translate>
              </v-list-item-title>
            </v-list-item>
          </template>

          <v-list-item v-show="$config.feature('files')" to="/index/files" class="nav-originals" @click.stop="">
            <v-list-item-title :class="`p-flex-menuitem menu-item ${rtl ? '--rtl' : ''}`">
              <translate key="Originals">Originals</translate>
            </v-list-item-title>
            <!-- TODO: fix filter -->
            <span v-show="config.count.files > 0 && canAccessPrivate" :class="`nav-count-item ${rtl ? '--rtl' : ''}`">{{ config.count.files }}</span>
            <!-- <span v-show="config.count.files > 0 && canAccessPrivate" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.files | abbreviateCount }}</span> -->
          </v-list-item>

          <v-list-item :to="{ name: 'hidden' }" class="nav-hidden" @click.stop="">
            <v-list-item-title :class="`p-flex-menuitem menu-item ${rtl ? '--rtl' : ''}`">
              <translate key="Hidden">Hidden</translate>
            </v-list-item-title>
            <!-- TODO: fix filter -->
            <span v-show="config.count.hidden > 0" :class="`nav-count-item ${rtl ? '--rtl' : ''}`">{{ config.count.hidden }}</span>
            <!-- <span v-show="config.count.hidden > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.hidden | abbreviateCount }}</span> -->
          </v-list-item>

          <v-list-item :to="{ name: 'errors' }" class="nav-errors" @click.stop="">
            <v-list-item-title :class="`p-flex-menuitem menu-item ${rtl ? '--rtl' : ''}`">
              <translate key="Errors">Errors</translate>
            </v-list-item-title>
          </v-list-item>
        </v-list-group>

        <template v-if="!config.disable.settings">
          <v-list-item v-if="isMini" v-show="$config.feature('settings')" :to="{ name: 'settings' }" class="nav-settings" @click.stop="">
            <v-icon class="ma-auto">mdi-cog</v-icon>
          </v-list-item>
          <v-list-group v-else-if="!isMini" v-show="$config.feature('settings')">
            <template #activator="{ props }">
              <v-list-item v-bind="props" :to="{ name: 'settings' }" class="nav-settings" @click.stop="">
                <v-list-item-title>
                  <v-icon>mdi-cog</v-icon>
                  <translate key="Settings" class="nav-item-title">Settings</translate>
                </v-list-item-title>
              </v-list-item>
            </template>

            <v-list-item v-if="canManageUsers" :to="{ path: '/admin/users' }" :exact="false" class="nav-admin-users" @click.stop="">
              <v-list-item-title :class="`menu-item ${rtl ? '--rtl' : ''}`">
                <translate>Users</translate>
              </v-list-item-title>
            </v-list-item>

            <v-list-item v-show="featFeedback" :to="{ name: 'feedback' }" :exact="true" class="nav-feedback" @click.stop="">
              <v-list-item-title :class="`menu-item ${rtl ? '--rtl' : ''}`">
                <translate>Feedback</translate>
              </v-list-item-title>
            </v-list-item>

            <v-list-item :to="{ name: 'license' }" :exact="true" class="nav-license" @click.stop="">
              <v-list-item-title :class="`menu-item ${rtl ? '--rtl' : ''}`">
                <translate key="License">License</translate>
              </v-list-item-title>
            </v-list-item>

            <v-list-item v-show="featUpgrade" :to="{ name: 'upgrade' }" class="nav-upgrade" :exact="true" @click.stop="">
              <v-list-item-title :class="`menu-item ${rtl ? '--rtl' : ''}`">
                <translate key="Upgrade">Upgrade</translate>
              </v-list-item-title>
            </v-list-item>

            <v-list-item :to="{ name: 'about' }" :exact="true" class="nav-about" @click.stop="">
              <v-list-item-title :class="`menu-item ${rtl ? '--rtl' : ''}`">
                <translate>About</translate>
              </v-list-item-title>
            </v-list-item>
          </v-list-group>
        </template>

        <v-list-item v-show="!auth" :to="{ name: 'login' }" class="nav-login" @click.stop="">
          <v-list-item-title>
            <v-icon>mdi-lock</v-icon>
            <translate key="Login">Login</translate>
          </v-list-item-title>
        </v-list-item>

        <router-link :to="{ name: 'upgrade' }" custom>
          <v-list-item v-show="featMembership" class="nav-upgrade" @click.stop="">
            <v-list-item-title v-if="isPro">
              <v-icon>mdi-check-circle</v-icon>
              <translate key="Upgrade" class="nav-item-title">Upgrade</translate>
            </v-list-item-title>
            <v-list-item-title v-else>
              <v-icon>mdi-diamond</v-icon>
              <translate key="Support Our Mission" class="nav-item-title">Support Our Mission</translate>
            </v-list-item-title>
          </v-list-item>
        </router-link>
      </v-list>

      <v-list class="p-user-box">
        <v-list-item v-show="$config.disconnected" to="/help/websockets" class="nav-connecting navigation" @click.stop="">
          <v-list-item-title class="text--warning">
            <v-icon color="warning">mdi-wifi-off</v-icon>
            <translate key="Offline" class="nav-item-title">Offline</translate>
          </v-list-item-title>
        </v-list-item>

        <v-list-item v-if="isMini && auth && !isPublic" class="nav-logout p-profile position-fixed bottom-0">
          <img :src="userAvatarURL" :alt="accountInfo" :title="accountInfo" class="rounded-circle" @click.stop="onAccountSettings" />
          <v-icon size="x-large" class="ma-auto mt-3 mb-3" @click.stop.prevent="onLogout">mdi-power</v-icon>
        </v-list-item>
        <v-list-item v-else-if="!isMini && auth && !isPublic" class="p-profile position-fixed bottom-0" bg-color="secondary" @click.stop="onAccountSettings">
          <div class="d-flex">
            <img :src="userAvatarURL" :alt="accountInfo" :title="accountInfo" class="rounded-circle" />

            <div class="ps-5">
              <v-list-item-title>{{ displayName }}</v-list-item-title>
              <v-list-item-subtitle>{{ accountInfo }}</v-list-item-subtitle>
            </div>

            <v-list-item-action class="ps-5" :title="$gettext('Logout')">
              <v-btn icon @click.stop.prevent="onLogout">
                <v-icon>mdi-power</v-icon>
              </v-btn>
            </v-list-item-action>
          </div>
        </v-list-item>
      </v-list>
    </v-navigation-drawer>
    <div id="mobile-menu" :class="{ active: speedDial }" @click.stop="speedDial = false">
      <div class="menu-content grow-top-right">
        <div class="menu-icons">
          <a v-if="auth && !isPublic" href="#" :title="$gettext('Logout')" class="menu-action nav-logout" @click.prevent="onLogout">
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
      <span v-if="config.legalUrl" class="clickable" @click.stop.prevent="onInfo()">{{ config.legalInfo }}</span>
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
    onAccountSettings: function () {
      if (this.$config.feature("account")) {
        this.$router.push({ name: "settings_account" });
      } else {
        this.$router.push({ name: "settings" });
      }
    },
    onInfo() {
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
