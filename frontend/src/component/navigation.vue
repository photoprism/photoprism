<template>
  <div id="p-navigation" :class="{'sidenav-visible': drawer}">
    <template v-if="visible && $vuetify.breakpoint.smAndDown">
      <v-toolbar dark fixed flat scroll-off-screen dense color="navigation darken-1" class="nav-small elevation-2"
                 @click.stop.prevent>
        <v-avatar class="nav-avatar" tile :size="28" :class="{'clickable': auth}" @click.stop.prevent="showNavigation()">
          <img :src="appIcon" :alt="config.name" :class="{'animate-hue': indexing}">
        </v-avatar>
        <v-toolbar-title class="nav-title">
          <span :class="{'clickable': auth}" @click.stop.prevent="showNavigation()">{{ page.title }}</span>
        </v-toolbar-title>
        <v-btn
            fab dark :ripple="false"
            color="transparent"
            class="mobile-menu-trigger elevation-0"
            @click.stop.prevent="speedDial = true"
        >
          <v-icon dark>more_vert</v-icon>
        </v-btn>
      </v-toolbar>
    </template>
    <template v-else-if="visible && !auth">
      <v-toolbar dark flat scroll-off-screen dense color="navigation darken-1" class="nav-small">
        <v-avatar class="nav-avatar" tile :size="28">
          <img :src="appIcon" :alt="config.name">
        </v-avatar>
        <v-toolbar-title class="nav-title">{{ page.title }}</v-toolbar-title>
        <v-btn
            fab dark :ripple="false"
            color="transparent"
            class="mobile-menu-trigger elevation-0"
            @click.stop.prevent="speedDial = true"
        >
          <v-icon dark>more_vert</v-icon>
        </v-btn>
      </v-toolbar>
    </template>
    <v-navigation-drawer
        v-if="visible && auth"
        v-model="drawer"
        :mini-variant="isMini"
        :width="270"
        :mobile-break-point="960"
        :mini-variant-width="80"
        class="nav-sidebar navigation p-flex-nav"
        fixed dark app
        :right="rtl"
    >
      <v-toolbar flat :dense="$vuetify.breakpoint.smAndDown">
        <v-list class="navigation-home">
          <v-list-tile class="nav-logo">
            <v-list-tile-avatar class="nav-avatar clickable" @click.stop.prevent="goHome">
              <img :src="appIcon" :alt="appName" :class="{'animate-hue': indexing}">
            </v-list-tile-avatar>
            <v-list-tile-content>
              <v-list-tile-title class="title">{{ appName }}</v-list-tile-title>
            </v-list-tile-content>
            <v-list-tile-action class="hidden-sm-and-down" :title="$gettext('Minimize')">
              <v-btn icon class="nav-minimize" @click.stop="toggleIsMini()">
                <v-icon v-if="!rtl">chevron_left</v-icon>
                <v-icon v-else>chevron_right</v-icon>
              </v-btn>
            </v-list-tile-action>
          </v-list-tile>
        </v-list>
      </v-toolbar>

      <v-list class="pt-3 p-flex-menu">
        <v-list-tile v-if="isMini && !isRestricted" class="nav-expand" @click.stop="toggleIsMini()">
          <v-list-tile-action :title="$gettext('Expand')">
            <v-icon v-if="!rtl">chevron_right</v-icon>
            <v-icon v-else>chevron_left</v-icon>
          </v-list-tile-action>
        </v-list-tile>

        <v-list-tile v-if="isMini && $config.feature('search')" to="/browse" class="nav-browse" @click.stop="">
          <v-list-tile-action :title="$gettext('Search')">
            <v-icon>search</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title>
              <translate>Search</translate>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>

        <v-list-group v-if="!isMini && $config.feature('search')" prepend-icon="search" no-action>
          <template #activator>
            <v-list-tile to="/browse" class="nav-browse" @click.stop="">
              <v-list-tile-content>
                <v-list-tile-title class="p-flex-menuitem">
                  <translate key="Search">Search</translate>
                  <span v-if="config.count.all > 0"
                        :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.all | abbreviateCount }}</span>
                </v-list-tile-title>
              </v-list-tile-content>
            </v-list-tile>
          </template>

          <v-list-tile :to="{name: 'browse', query: { q: 'mono:true quality:3 photo:true' }}" :exact="true"
                       class="nav-monochrome" @click.stop="">
            <v-list-tile-content>
              <v-list-tile-title :class="`p-flex-menuitem menu-item ${rtl ? '--rtl' : ''}`">
                <translate>Monochrome</translate>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>

          <v-list-tile :to="{name: 'browse', query: { q: 'panoramas' }}" :exact="true" class="nav-panoramas"
                       @click.stop="">
            <v-list-tile-content>
              <v-list-tile-title :class="`menu-item ${rtl ? '--rtl' : ''}`">
                <translate>Panoramas</translate>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>

          <v-list-tile :to="{name: 'browse', query: { q: 'animated' }}" :exact="true" class="nav-animated"
                       @click.stop="">
            <v-list-tile-content>
              <v-list-tile-title :class="`menu-item ${rtl ? '--rtl' : ''}`">
                <translate>Animated</translate>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>

          <v-list-tile v-show="isSponsor" :to="{name: 'browse', query: { q: 'vectors' }}" :exact="true" class="nav-vectors"
                       @click.stop="">
            <v-list-tile-content>
              <v-list-tile-title :class="`menu-item ${rtl ? '--rtl' : ''}`">
                <translate>Vectors</translate>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>

          <v-list-tile :to="{name: 'photos', query: { q: 'stacks' }}" :exact="true" class="nav-stacks" @click.stop="">
            <v-list-tile-content>
              <v-list-tile-title :class="`p-flex-menuitem menu-item ${rtl ? '--rtl' : ''}`">
                <translate>Stacks</translate>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>

          <v-list-tile :to="{name: 'photos', query: { q: 'scans' }}" :exact="true" class="nav-scans" @click.stop="">
            <v-list-tile-content>
              <v-list-tile-title :class="`p-flex-menuitem menu-item ${rtl ? '--rtl' : ''}`">
                <translate>Scans</translate>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>

          <v-list-tile v-if="canManagePhotos" v-show="$config.feature('review')" to="/review" class="nav-review"
                       @click.stop="">
            <v-list-tile-content>
              <v-list-tile-title :class="`p-flex-menuitem menu-item ${rtl ? '--rtl' : ''}`">
                <translate>Review</translate>
                <span v-show="config.count.review > 0"
                      :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.review | abbreviateCount }}</span>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>

          <v-list-tile v-show="$config.feature('archive')" to="/archive" class="nav-archive" @click.stop="">
            <v-list-tile-content>
              <v-list-tile-title :class="`p-flex-menuitem menu-item ${rtl ? '--rtl' : ''}`">
                <translate>Archive</translate>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>
        </v-list-group>

        <v-list-tile v-if="isMini" v-show="$config.feature('albums')" to="/albums" class="nav-albums" @click.stop="">
          <v-list-tile-action :title="$gettext('Albums')">
            <v-icon>bookmark</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title :class="`menu-item ${rtl ? '--rtl' : ''}`">
              <translate key="Albums">Albums</translate>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>

        <v-list-group v-if="!isMini" v-show="$config.feature('albums')" prepend-icon="bookmark" no-action>
          <template #activator>
            <v-list-tile :to="{ name: 'albums' }" class="nav-albums" @click.stop="">
              <v-list-tile-content>
                <v-list-tile-title class="p-flex-menuitem">
                  <translate key="Albums">Albums</translate>
                  <span v-if="config.count.albums > 0"
                        :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.albums | abbreviateCount }}</span>
                </v-list-tile-title>
              </v-list-tile-content>
            </v-list-tile>
          </template>

          <v-list-tile to="/unsorted" class="nav-unsorted">
            <v-list-tile-content>
              <v-list-tile-title :class="`p-flex-menuitem menu-item ${rtl ? '--rtl' : ''}`">
                <translate key="Unsorted">Unsorted</translate>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>
        </v-list-group>

        <v-list-tile v-if="isMini && $config.feature('videos')" to="/videos" class="nav-video" @click.stop="">
          <v-list-tile-action :title="$gettext('Videos')">
            <v-icon>play_circle_fill</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title :class="`menu-item ${rtl ? '--rtl' : ''}`">
              <translate key="Videos">Videos</translate>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>

        <v-list-group v-if="!isMini && $config.feature('videos')" prepend-icon="play_circle_fill" no-action>
          <template #activator>
            <v-list-tile to="/videos" class="nav-video" @click.stop="">
              <v-list-tile-content>
                <v-list-tile-title class="p-flex-menuitem">
                  <translate key="Videos">Videos</translate>
                  <span v-show="config.count.videos > 0"
                        :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.videos | abbreviateCount }}</span>
                </v-list-tile-title>
              </v-list-tile-content>
            </v-list-tile>
          </template>

          <v-list-tile :to="{name: 'live'}" class="nav-live" @click.stop="">
            <v-list-tile-content>
              <v-list-tile-title :class="`p-flex-menuitem menu-item ${rtl ? '--rtl' : ''}`">
                <translate>Live Photos</translate>
                <span v-show="config.count.live > 0"
                      :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.live | abbreviateCount }}</span>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>
        </v-list-group>

        <v-list-tile v-show="$config.feature('people') && (canManagePeople || config.count.people > 0)" :to="{ name: 'people' }" class="nav-people" @click.stop="">
          <v-list-tile-action :title="$gettext('People')">
            <v-icon>person</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title class="p-flex-menuitem">
              <translate key="People">People</translate>
              <span v-show="config.count.people > 0"
                    :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.people | abbreviateCount }}</span>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>

        <v-list-tile v-show="$config.feature('favorites')" :to="{ name: 'favorites' }" class="nav-favorites" @click.stop="">
          <v-list-tile-action :title="$gettext('Favorites')">
            <v-icon>favorite</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title class="p-flex-menuitem">
              <translate key="Favorites">Favorites</translate>
              <span v-show="config.count.favorites > 0"
                    :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.favorites | abbreviateCount }}</span>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>

        <v-list-tile v-show="$config.feature('moments')" :to="{ name: 'moments' }" class="nav-moments"
                     @click.stop="">
          <v-list-tile-action :title="$gettext('Moments')">
            <v-icon>star</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title class="p-flex-menuitem">
              <translate key="Moments">Moments</translate>
              <span v-show="config.count.moments > 0"
                    :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.moments | abbreviateCount }}</span>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>

        <v-list-tile v-show="$config.feature('moments')" :to="{ name: 'calendar' }" class="nav-calendar" @click.stop="">
          <v-list-tile-action :title="$gettext('Calendar')">
            <v-icon>date_range</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title class="p-flex-menuitem">
              <translate key="Calendar">Calendar</translate>
              <span v-show="config.count.months > 0"
                    :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.months | abbreviateCount }}</span>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>

        <v-list-tile v-if="isRestricted" v-show="$config.feature('places')" to="/states" class="nav-states" @click.stop="">
          <v-list-tile-action :title="$gettext('States')">
            <v-icon>near_me</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title class="p-flex-menuitem" @click.stop="">
              <translate key="States">States</translate>
              <span v-show="config.count.states > 0"
                    :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.states | abbreviateCount }}</span>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>

        <template v-if="canSearchPlaces">
        <v-list-tile v-if="isMini" v-show="canSearchPlaces && $config.feature('places')" :to="{ name: 'places' }" class="nav-places"
                     @click.stop="">
          <v-list-tile-action :title="$gettext('Places')">
            <v-icon>place</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title class="p-flex-menuitem">
              <translate key="Places">Places</translate>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>

        <v-list-group v-if="!isMini" v-show="canSearchPlaces && $config.feature('places')" prepend-icon="place" no-action>
          <template #activator>
            <v-list-tile to="/places" class="nav-places" @click.stop="">
              <v-list-tile-content>
                <v-list-tile-title class="p-flex-menuitem">
                  <translate key="Places">Places</translate>
                  <span v-show="config.count.places > 0"
                        :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.places | abbreviateCount }}</span>
                </v-list-tile-title>
              </v-list-tile-content>
            </v-list-tile>
          </template>

          <v-list-tile to="/states" class="nav-states" @click.stop="">
            <v-list-tile-content>
              <v-list-tile-title :class="`p-flex-menuitem menu-item ${rtl ? '--rtl' : ''}`">
                <translate key="States">States</translate>
                <span v-show="config.count.states > 0"
                      :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.states | abbreviateCount }}</span>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>
        </v-list-group>
        </template>

        <v-list-tile v-show="$config.feature('labels')" to="/labels" class="nav-labels" @click.stop="">
          <v-list-tile-action :title="$gettext('Labels')">
            <v-icon>label</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title class="p-flex-menuitem">
              <translate key="Labels">Labels</translate>
              <span v-show="config.count.labels > 0"
                    :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.labels | abbreviateCount }}</span>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>

        <v-list-tile v-show="$config.feature('folders')" :to="{ name: 'folders' }" class="nav-folders" @click.stop="">
          <v-list-tile-action :title="$gettext('Folders')">
            <v-icon>folder</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title class="p-flex-menuitem">
              <translate key="Folders">Folders</translate>
              <span v-show="config.count.folders > 0"
                    :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.folders | abbreviateCount }}</span>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>

        <v-list-tile v-show="$config.feature('private')" to="/private" class="nav-private" @click.stop="">
          <v-list-tile-action :title="$gettext('Private')">
            <v-icon>lock</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title class="p-flex-menuitem">
              <translate key="Private">Private</translate>
              <span v-show="config.count.private > 0"
                    :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.private | abbreviateCount }}</span>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>

        <v-list-tile v-if="isMini && $config.feature('library')" :to="{ name: 'library_index' }" class="nav-library" @click.stop="">
          <v-list-tile-action :title="$gettext('Library')">
            <v-icon>camera_roll</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title>
              <translate key="Library">Library</translate>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>

        <v-list-group v-if="!isMini && $config.feature('library')" prepend-icon="camera_roll" no-action>
          <template #activator>
            <v-list-tile :to="{ name: 'library_index' }" class="nav-library" @click.stop="">
              <v-list-tile-content>
                <v-list-tile-title class="p-flex-menuitem">
                  <translate key="Library">Library</translate>
                </v-list-tile-title>
              </v-list-tile-content>
            </v-list-tile>
          </template>

          <v-list-tile v-show="$config.feature('files')" to="/index/files" class="nav-originals" @click.stop="">
            <v-list-tile-content>
              <v-list-tile-title :class="`p-flex-menuitem menu-item ${rtl ? '--rtl' : ''}`">
                <translate key="Originals">Originals</translate>
                <span v-show="config.count.files > 0 && canAccessPrivate"
                      :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.files | abbreviateCount }}</span>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>

          <v-list-tile :to="{ name: 'hidden' }" class="nav-hidden" @click.stop="">
            <v-list-tile-content>
              <v-list-tile-title :class="`p-flex-menuitem menu-item ${rtl ? '--rtl' : ''}`">
                <translate key="Hidden">Hidden</translate>
                <span v-show="config.count.hidden > 0"
                      :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.hidden | abbreviateCount }}</span>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>

          <v-list-tile :to="{ name: 'errors' }" class="nav-errors" @click.stop="">
            <v-list-tile-content>
              <v-list-tile-title :class="`p-flex-menuitem menu-item ${rtl ? '--rtl' : ''}`">
                <translate key="Errors">Errors</translate>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>
        </v-list-group>

        <template v-if="!config.disable.settings">
          <v-list-tile v-if="isMini" v-show="$config.feature('settings')" :to="{ name: 'settings' }" class="nav-settings" @click.stop="">
            <v-list-tile-action :title="$gettext('Settings')">
              <v-icon>settings</v-icon>
            </v-list-tile-action>

            <v-list-tile-content>
              <v-list-tile-title>
                <translate key="Settings">Settings</translate>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>

          <v-list-group v-else v-show="$config.feature('settings')" prepend-icon="settings" no-action>
            <template #activator>
              <v-list-tile :to="{ name: 'settings' }" class="nav-settings" @click.stop="">
                <v-list-tile-content>
                  <v-list-tile-title>
                    <translate key="Settings">Settings</translate>
                  </v-list-tile-title>
                </v-list-tile-content>
              </v-list-tile>
            </template>

            <v-list-tile :to="{ name: 'about' }" :exact="true" class="nav-about" @click.stop="">
              <v-list-tile-content>
                <v-list-tile-title :class="`menu-item ${rtl ? '--rtl' : ''}`">
                  <translate>About</translate>
                </v-list-tile-title>
              </v-list-tile-content>
            </v-list-tile>

            <v-list-tile v-if="canManageUsers" :to="{ path: '/admin/users' }" :exact="false" class="nav-admin-users" @click.stop="">
              <v-list-tile-content>
                <v-list-tile-title :class="`menu-item ${rtl ? '--rtl' : ''}`">
                  <translate>Users</translate>
                </v-list-tile-title>
              </v-list-tile-content>
            </v-list-tile>

            <v-list-tile v-show="!isPublic && isAdmin && isSponsor" :to="{ name: 'feedback' }" :exact="true" class="nav-feedback"
                         @click.stop="">
              <v-list-tile-content>
                <v-list-tile-title :class="`menu-item ${rtl ? '--rtl' : ''}`">
                  <translate>Feedback</translate>
                </v-list-tile-title>
              </v-list-tile-content>
            </v-list-tile>

            <v-list-tile :to="{ name: 'license' }" :exact="true" class="nav-license" @click.stop="">
              <v-list-tile-content>
                <v-list-tile-title :class="`menu-item ${rtl ? '--rtl' : ''}`">
                  <translate key="License">License</translate>
                </v-list-tile-title>
              </v-list-tile-content>
            </v-list-tile>

            <v-list-tile v-show="isAdmin && !isPublic && !isDemo && featUpgrade" :to="{ name: 'upgrade' }" class="nav-upgrade" :exact="true" @click.stop="">
              <v-list-tile-content>
                <v-list-tile-title :class="`menu-item ${rtl ? '--rtl' : ''}`">
                  <translate key="Upgrade">Upgrade</translate>
                </v-list-tile-title>
              </v-list-tile-content>
            </v-list-tile>
          </v-list-group>
        </template>

        <v-list-tile v-show="!auth" :to="{ name: 'login' }" class="nav-login" @click.stop="">
          <v-list-tile-action :title="$gettext('Login')">
            <v-icon>lock</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title>
              <translate key="Login">Login</translate>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>
      </v-list>

      <v-list class="p-user-box">

        <v-list-tile v-show="$config.disconnected" to="/help/websockets" class="nav-connecting navigation"
                     @click.stop="">
          <v-list-tile-action :title="$gettext('Offline')">
            <v-icon color="warning">wifi_off</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title class="text--warning">
              <translate key="Offline">Offline</translate>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>

        <v-list-tile v-show="auth && !isPublic && $config.feature('settings')" class="p-profile" @click.stop="onAccount">
          <v-list-tile-avatar size="36">
            <img :src="userAvatarURL" :alt="accountInfo" :title="accountInfo">
          </v-list-tile-avatar>

          <v-list-tile-content>
            <v-list-tile-title>{{ displayName }}</v-list-tile-title>
            <v-list-tile-sub-title>{{ accountInfo }}</v-list-tile-sub-title>
          </v-list-tile-content>

          <v-list-tile-action :title="$gettext('Logout')">
            <v-btn icon @click.stop.prevent="onLogout">
              <v-icon>power_settings_new</v-icon>
            </v-btn>
          </v-list-tile-action>
        </v-list-tile>

        <v-list-tile v-show="isMini && auth && !isPublic" class="nav-logout" @click.stop.prevent="onLogout">
          <v-list-tile-action :title="$gettext('Logout')">
            <v-icon>power_settings_new</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title>
              <translate key="Logout">Logout</translate>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>
      </v-list>
    </v-navigation-drawer>
    <div id="mobile-menu" :class="{'active': speedDial}" @click.stop="speedDial = false">
      <div class="menu-content grow-top-right">
        <div class="menu-icons">
          <a v-if="auth && !isPublic" href="#" :title="$gettext('Logout')" class="menu-action nav-logout"
             @click.prevent="onLogout">
            <v-icon>power_settings_new</v-icon>
          </a>
          <a href="#" :title="$gettext('Reload')" class="menu-action nav-reload" @click.prevent="reloadApp">
            <v-icon>refresh</v-icon>
          </a>
          <router-link v-if="auth && $config.feature('account')"
                       :to="{ name: 'settings_account' }" :title="$gettext('Account')" class="menu-action nav-account">
            <v-icon>admin_panel_settings</v-icon>
          </router-link>
          <router-link v-if="auth && $config.feature('settings') && !routeName('settings')" :to="{ name: 'settings' }"
                       :title="$gettext('Settings')" class="menu-action nav-settings">
            <v-icon>settings</v-icon>
          </router-link>
          <a v-if="auth && !config.readonly && $config.feature('upload')" href="#" :title="$gettext('Upload')"
             class="menu-action nav-upload" @click.prevent="openUpload()">
            <v-icon>cloud_upload</v-icon>
          </a>
          <router-link v-if="!auth && !isPublic" :to="{ name: 'login' }" :title="$gettext('Login')"
                       class="menu-action nav-login">
            <v-icon>login</v-icon>
          </router-link>
        </div>
        <div class="menu-actions">
          <div v-if="auth && !routeName('browse')&& $config.feature('search')" class="menu-action nav-search">
            <router-link to="/browse">
              <v-icon>search</v-icon>
              <translate>Search</translate>
            </router-link>
          </div>
          <div v-if="auth && !routeName('albums') && $config.feature('albums')" class="menu-action nav-albums">
            <router-link to="/albums">
              <v-icon>bookmark</v-icon>
              <translate>Albums</translate>
            </router-link>
          </div>
          <div v-if="auth && canManagePeople && !routeName('people') && $config.feature('people')" class="menu-action nav-people">
            <router-link to="/people">
              <v-icon>person</v-icon>
              <translate>People</translate>
            </router-link>
          </div>
          <div v-if="auth && canSearchPlaces && !routeName('places') && $config.feature('places')" class="menu-action nav-places">
            <router-link to="/places">
              <v-icon>place</v-icon>
              <translate>Places</translate>
            </router-link>
          </div>
          <div v-if="auth && !routeName('files') && $config.feature('files') && $config.feature('library')"
               class="menu-action nav-files">
            <router-link to="/index/files">
              <v-icon>folder</v-icon>
              <translate>Files</translate>
            </router-link>
          </div>
          <div v-if="auth && !routeName('library_index') && $config.feature('library')" class="menu-action nav-index">
            <router-link :to="{ name: 'library_index' }">
              <v-icon>camera_roll</v-icon>
              <translate>Index</translate>
            </router-link>
          </div>
          <div v-if="auth && !routeName('index') && $config.feature('library') && $config.feature('logs')" class="menu-action nav-logs">
            <router-link :to="{ name: 'library_logs' }">
              <v-icon>feed</v-icon>
              <translate>Logs</translate>
            </router-link>
          </div>
          <div v-if="!isPublic && !isSponsor && isAdmin" class="menu-action nav-membership">
            <router-link :to="{ name: 'upgrade' }">
              <v-icon>diamond</v-icon>
              <translate>Upgrade</translate>
            </router-link>
          </div>
          <div class="menu-action nav-manual"><a href="https://link.photoprism.app/docs" target="_blank">
            <v-icon>auto_stories</v-icon>
            <translate>User Guide</translate>
          </a></div>
          <div v-if="config.legalUrl && isSponsor" class="menu-action nav-legal"><a :href="config.legalUrl"
                                                                                      target="_blank">
            <v-icon>info</v-icon>
            <translate>Legal Information</translate>
          </a></div>
        </div>
      </div>
    </div>
    <div v-if="config.legalInfo && visible" id="legal-info">
      <a v-if="config.legalUrl" :href="config.legalUrl" target="_blank">{{ config.legalInfo }}</a>
      <span v-else>{{ config.legalInfo }}</span>
    </div>
    <p-reload-dialog :show="reload.dialog" @close="reload.dialog = false"></p-reload-dialog>
    <p-upload-dialog :show="upload.dialog" @cancel="upload.dialog = false"
                     @confirm="upload.dialog = false"></p-upload-dialog>
    <p-photo-edit-dialog :show="edit.dialog" :selection="edit.selection" :index="edit.index" :album="edit.album"
                         @close="edit.dialog = false"></p-photo-edit-dialog>
  </div>
</template>

<script>
import Album from "model/album";
import Event from "pubsub-js";

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
    }
  },
  data() {
    const appName = this.$config.getName();

    let appNameSuffix = "";
    let appNameParts = appName.split(" ");

    if (appNameParts.length > 1) {
      appNameSuffix = appNameParts.slice(1, 9).join(" ");
    }

    const isRestricted = this.$config.deny("photos", "access_library");

    return {
      canSearchPlaces: this.$config.allow("places", "search"),
      canAccessPrivate: !isRestricted && this.$config.allow("photos", "access_private"),
      canManagePhotos: this.$config.allow("photos", "manage"),
      canManagePeople: this.$config.allow("people", "manage"),
      canManageUsers: this.$config.allow("users", "manage"),
      appNameSuffix: appNameSuffix,
      appName: this.$config.getName(),
      appAbout: this.$config.getAbout(),
      appIcon: this.$config.getIcon(),
      indexing: false,
      drawer: null,
      featUpgrade: this.$config.getLicense() === "ce",
      isRestricted: isRestricted,
      isMini: localStorage.getItem('last_navigation_mode') !== 'false' || isRestricted,
      isPublic: this.$config.get("public"),
      isDemo: this.$config.get("demo"),
      isAdmin: this.$session.isAdmin(),
      isSponsor: this.$config.isSponsor(),
      isTest: this.$config.test,
      isReadOnly: this.$config.get("readonly"),
      session: this.$session,
      config: this.$config.values,
      page: this.$config.page,
      user: this.$session.getUser(),
      reload: {
        dialog: false,
      },
      upload: {
        dialog: false,
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
      return this.$session.getUser().getAvatarURL('tile_50');
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
    this.subscriptions.push(Event.subscribe('index', this.onIndex));
    this.subscriptions.push(Event.subscribe('import', this.onIndex));
    this.subscriptions.push(Event.subscribe("dialog.reload", () => this.reload.dialog = true));
    this.subscriptions.push(Event.subscribe("dialog.upload", () => this.upload.dialog = true));
    this.subscriptions.push(Event.subscribe("dialog.edit", (ev, data) => {
      if (!this.edit.dialog) {
        this.edit.dialog = true;
        this.edit.index = data.index;
        this.edit.selection = data.selection;
        this.edit.album = data.album;
      }
    }));
  },
  destroyed() {
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
      if (this.auth && !this.isReadOnly && this.$config.feature('upload')) {
        this.upload.dialog = true;
      } else {
        this.goHome();
      }
    },
    goHome() {
      if (this.$route.name !== "home") {
        this.$router.push({name: "home"});
      }
    },
    showNavigation() {
      if (this.auth) {
        this.drawer = true;
        this.isMini = this.isRestricted;
      }
    },
    createAlbum() {
      let name = "New Album";
      const album = new Album({Title: name, Favorite: false});
      album.save();
    },
    toggleIsMini() {
      this.isMini = !this.isMini;
      localStorage.setItem('last_navigation_mode', `${this.isMini}`);
    },
    onAccount: function () {
      this.$router.push({name: "settings_account"});
    },
    onLogout() {
      this.$session.logout();
    },
    onIndex(ev) {
      if (!ev) {
        return;
      }

      const type = ev.split('.')[1];

      switch (type) {
        case "file":
        case "folder":
        case "indexing":
          this.indexing = true;
          break;
        case 'completed':
          this.indexing = false;
          break;
      }
    },
  },
};
</script>
