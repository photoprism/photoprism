---
name: Bug Report
about: Report a new and clearly identified bug that must be fixed directly in the application
title: 'SHORT DESCRIPTION OF THE PROBLEM YOU ARE REPORTING'
labels: bug
assignees: ''

---

PLEASE PROCEED ONLY IF YOU ARE ABSOLUTELY SURE THAT THIS IS NOT A TECHNICAL SUPPORT INCIDENT AND/OR POSSIBLY A PROBLEM WITH SOME OTHER SOFTWARE YOU ARE USING. VISIT <https://www.photoprism.app/kb/getting-support> TO LEARN MORE ABOUT OUR SUPPORT OPTIONS. THANK YOU FOR YOUR CAREFUL CONSIDERATION!

---------------------------------------------------------------------------

#### 1. What is not working as documented?

Be as specific as possible and explain which part of the software is not [working as documented](https://docs.photoprism.app/), e.g. "image not found" or "wrong thumbnail" would not be detailed enough.

Links to the related documentation on [docs.photoprism.app](https://docs.photoprism.app/):
- ...

*Please never report [known issues](https://docs.photoprism.app/known-issues/) or [missing features](https://github.com/photoprism/photoprism/issues) as bugs, and do not submit bug reports for the purpose of getting [technical support](https://www.photoprism.app/kb/getting-support) or because you have not received a response in our [public community forums](https://github.com/photoprism/photoprism/discussions). Thank you very much!*

#### 2. How can we reproduce it?

Steps to reproduce the behavior:

1. Go to '...'
2. Click on '....'
3. Scroll down to '....'
4. See error

When reporting an import, indexing, or performance issue, please include the number and type of pictures in your library, as well as any configuration options you have changed, such as for thumbnail quality.

#### 3. What behavior do you expect?

Give us a clear and concise description of what you expect.

#### 4. What could be the cause of your problem?

Always try to determine the cause of your problem using the checklists at <https://docs.photoprism.app/getting-started/troubleshooting/> before submitting a bug report.

#### 5. Can you provide us with example files for testing, error logs, or screenshots?

Please include sample files or screenshots that help to reproduce your problem. You can also email files or share a download link, see <https://www.photoprism.app/contact> for details.

Visit <https://docs.photoprism.app/getting-started/troubleshooting/browsers/> to learn how to diagnose frontend issues.

**Important: If it is an import, indexing or metadata issue, we require sample files and logs from you.** Otherwise, we will not be able to process your report. If it is an import problem specifically, please always provide us with an archive of the files before you imported them so we can reproduce the behavior.

#### 6. Which software versions do you use?

(a) PhotoPrism Architecture & Build Number: AMD64, ARM64, ARMv7,...

(b) Database Type & Version: MariaDB, MySQL, SQLite,...

(c) Operating System Types & Versions: Linux, Windows, Android,...

(d) Browser Types & Versions: Firefox, Chrome, Safari on iPhone,...

(e) Ad Blockers, Browser Plugins, and/or Firewall Software?

You can find the version/build number of the app in *Settings* by scrolling to the bottom. Note that MySQL 8 support has been discontinued, see system requirements at <https://docs.photoprism.app/getting-started/#system-requirements>.

*Always provide database and operating system details if it is a backend, import, or indexing issue. Should it be a frontend issue, at a minimum we require you to provide web browser and operating system details. When reporting a performance problem, we ask that you provide us with complete information about your environment, as there may be more than one cause.*

#### 7. On what kind of device is PhotoPrism installed?

This is especially important if you are reporting a performance, import, or indexing issue. You can skip this if you're reporting a problem you found in our public demo, or if it's a completely unrelated issue, such as incorrect page layout.

(a) Device / Processor Type: Raspberry Pi 4, Intel Core i7-3770, AMD Ryzen 7 3800X,...

(b) Physical Memory & Swap Space in GB

(c) Storage Type: HDD, SSD, RAID, USB, Network Storage,...

(d) Anything else that might be helpful to know?

*Always provide device, memory, and storage details if you have a backend, performance, import, or indexing issue.*

#### 8. Do you use a Reverse Proxy, Firewall, VPN, or CDN?

If yes, please specify type and version. You can skip this if you are reporting a completely unrelated issue.

*Always provide this information when you have a reliability, performance, or frontend problem, such as failed uploads, connection errors, broken thumbnails, or video playback issues.*

**Using NGINX?** Please also provide the configuration and/or consider asking the NGINX community for advice as we do not specialize in supporting their product. Docs can be found at <https://docs.photoprism.app/getting-started/proxies/nginx/>.
