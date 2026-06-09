// Taken from https://testcafe.io/documentation/404388/guides/advanced-guides/modify-reporter-output
// And updated to show h m s ms time format.
function onBeforeWriteHook(writeInfo) { // This function will fire every time the reporter calls the "write" method.
    if (writeInfo.initiator === 'reportTestDone') { // The "initiator" property contains the name of the reporter event that triggered the hook.
         const {
            name,
            testRunInfo,
            meta
        } = writeInfo.data || {}; // If you attached this hook to a compatible reporter (such as "spec" or "list"), the hook can process data related to the event.
        const testDuration = new Date(testRunInfo.durationMs).toISOString().slice(11, -1); // Save the duration of the test.
        writeInfo.formattedText = writeInfo.formattedText + ' (' + testDuration + ')'; // Add test duration to the reporter output.
    };
}


module.exports = { // Attach the hook
    hooks: {
        reporter: {
            onBeforeWrite: {
                'spec': onBeforeWriteHook, // This hook will fire when you use the default "spec" reporter.
            },
        },
    },
    skipJsErrors: true,
    quarantineMode: false,
    selectorTimeout: 3500,
    assertionTimeout: 3500,
    hostname: "localhost",
    retryTestPages: true,
    screenshots: {
        path: "tests/acceptance/screenshots/",
        takeOnFails: true
    },
};