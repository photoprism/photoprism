<template>
  <div class="p-page p-page-about">
    <v-toolbar flat color="secondary">
      <v-toolbar-title>
        <translate>Whatever it is, we'd love to hear from you!</translate>
      </v-toolbar-title>

      <v-spacer></v-spacer>

      <v-btn icon href="https://github.com/photoprism/photoprism" target="_blank" class="action-github" title="GitHub">
        <img src="/static/brands/github.svg" width="24" alt="GitHub">
      </v-btn>
    </v-toolbar>
    <v-container fluid class="pa-4">
      <p class="body-1">Feel free to use the
        <a target="_blank" href="https://gitter.im/browseyourlife/community" class="highlight">chat</a>
        or send an e-mail to <a target="_blank" href="mailto:hello@photoprism.org" class="highlight">hello@photoprism.org</a>
        if you have questions, want to support our work, or just want to say hello.</p>

      <p>
        <router-link to="/about/license">
          <img src="/static/img/badge-agpl.svg" alt="License AGPL v3" style="max-width:100%;"/>
        </router-link>
        <a target="_blank" href="https://docs.photoprism.org/" rel="nofollow"><img src="/static/img/badge-docs.svg"
                                                                                   alt="Documentation"
                                                                                   style="max-width:100%;"></a>
        <a target="_blank" href="https://gitter.im/browseyourlife/community" rel="nofollow"><img
                src="/static/img/badge-chat.svg" alt="Community Chat" style="max-width:100%;"></a>
        <a target="_blank" href="https://twitter.com/browseyourlife" rel="nofollow"><img
                src="/static/img/badge-twitter.svg" alt="Twitter" style="max-width:100%;"></a>
      </p>

      <h2 class="py-2 subheading"><translate>Who we are</translate></h2>

      <h3 class="py-2 body-2">Andrea Ceroni</h3>
      <p class="body-1">Andrea is a data scientist specialized in temporal information retrieval and machine
        learning.
        He holds a PhD in Computer Science from the Leibniz University of Hannover (L3S Research Center) and wrote
        numerous papers on topics such as
        <a target="_blank" href="https://www.researchgate.net/profile/Andrea_Ceroni/publication/323222448_Personal_Photo_Management_and_Preservation/links/5a995f8da6fdcc3cbac8fa59/Personal-Photo-Management-and-Preservation.pdf" class="highlight">Personal
          Photo Management and Preservation</a>
        and <a target="_blank" href="https://www.iti.gr/~bmezaris/publications/hmmp@icme2015_2_preprint.pdf" class="highlight">Photo Selection Models for
          Personal Photo Collections</a>.</p>
      <p class="body-1">You can find him on <a target="_blank"
              href="https://scholar.google.de/citations?user=JHsQY5YAAAAJ&amp;hl=en" class="highlight">Google Scholar</a> and <a target="_blank"
              href="https://www.linkedin.com/in/andrea-ceroni/" class="highlight">LinkedIn</a></p>

      <h3 class="py-2 body-2">Theresa Gresch</h3>
      <p class="body-1">Theresa works as a freelance product manager and developer in Berlin. She has a Master's in
        Neurobiology
        and aims to dive deeper into machine learning while working on this project.</p>
      <p class="body-1">You can find her on <a target="_blank" href="https://github.com/graciousgrey" class="highlight">GitHub</a> and <a
              target="_blank" href="https://www.linkedin.com/in/theresa-gresch-886924103/" class="highlight">LinkedIn</a></p>

      <h3 class="py-2 body-2">Michael Mayer</h3>
      <p class="body-1">Michael learned coding on an Atari 1040 ST and started his first open-source projects in
        the 90s. He has more than 20 years of experience in building Web applications.
        His motivation is to explore the latest technologies and build an amazing product outside the constraints of a
        corporate environment.</p>
      <p class="body-1">You can find him on <a target="_blank" href="https://github.com/lastzero" class="highlight">GitHub</a>,
        <a target="_blank" href="https://www.linkedin.com/in/lastzero/" class="highlight">LinkedIn</a>
        and <a target="_blank" href="https://twitter.com/lastzero" class="highlight">Twitter</a>
      </p>

      <h3 class="py-2 body-2">Guy Sheffer</h3>
      <p class="body-1">Known as <a target="_blank" href="https://github.com/guysoft">GuySoft</a> on the web. Active
        developer in
        the Free Software and Maker community.
        Creator of <a target="_blank" href="https://github.com/guysoft/OctoPi" class="highlight">OctoPi</a> and
        <a target="_blank" href="https://github.com/guysoft/FullPageOS" class="highlight">FullPageOS</a>,
        which have hundreds of thousands of downloads. Raspberry Pi distro expert.
        Currently VP R&amp;D and Co-Founder at <a target="_blank" href="https://shapedo.com/" class="highlight">ShapeDo</a>.</p>

      <h3 class="pb-1 body-2">...and many more</h3>
      <p class="body-1">Thank you to
        <a target="_blank" href="https://github.com/photoprism/photoprism/graphs/contributors" class="highlight">everyone who
        contributed</a> to this project!</p>


      <h2 class="py-2 subheading"><translate>Trademarks</translate></h2>
      <p class="body-1">
        <translate>PhotoPrism™ is a registered trademark of Michael Mayer.</translate>
        <translate>You may use it as required to describe our software, run your own server, for educational purposes, but not for offering commercial goods, products, or services without prior written permission. In other words, please ask.</translate>
      </p>
    </v-container>

    <p-about-footer></p-about-footer>
  </div>
</template>

<script>
    export default {
        name: 'p-page-about',
        data() {
            return {};
        },
        methods: {},
    };
</script>
