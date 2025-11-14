# Export Universal APK

[![Step changelog](https://shields.io/github/v/release/bitrise-steplib/bitrise-step-export-universal-apk?include_prereleases&label=changelog&color=blueviolet)](https://github.com/bitrise-steplib/bitrise-step-export-universal-apk/releases)

Exports a universal APK from an Android App Bundle.


<details>
<summary>Description</summary>

This Step generates a universal APK from an Android App Bundle, and exports it to the `$BITRISE_APK_PATH` Environment Variable so that the next Step or [Ship add-on](https://devcenter.bitrise.io/deploy/ship/) can pick it up. The Step also signs the generated APK with the keystore file you uploaded to the [Code Signing](https://devcenter.bitrise.io/code-signing/android-code-signing/android-code-signing-index/) tab or, if there was no keystore available, it signs the APK with a debug keystore file.

### Configuring the Step
1. Insert the Step after a build Step in your Workflow.
2. The **Android App Bundle path** input field is automatically filled out by the output of the previous build Step.
3. The **Keystore URL** is automatically filled out based on the uploaded keystore file on the **Code Signing** tab.
4. If the keystore file is uploaded to the **Code Signing** tab, the **Keystore alias**, **Keystore password**, and **Private key password** inputs are automatically populated.
5. The latest Bundletool version is set in the respective input. If, for any reason, you wish to use an older version, you can add it here, but make sure you use the [correct version](https://github.com/google/bundletool/releases).

### Troubleshooting
This Step works with Bundletool's latest version which is automatically set in the respective Step input. If you wish to switch to an older version, you have to add it manually. Make sure you add the [correct version](https://github.com/google/bundletool/releases), otherwise the Step will fail.

### Useful links
- [Android code signing](https://devcenter.bitrise.io/code-signing/android-code-signing/android-code-signing-index/)
- [Deploying and Android app](https://devcenter.bitrise.io/deploy/android-deploy/android-deployment-index/)

### Related Steps
- [Android Sign](https://www.bitrise.io/integrations/steps/sign-apk)
- [Android Build](https://www.bitrise.io/integrations/steps/android-build)

</details>

## 🧩 Get started

Add this step directly to your workflow in the [Bitrise Workflow Editor](https://docs.bitrise.io/en/bitrise-ci/workflows-and-pipelines/steps/adding-steps-to-a-workflow.html).

You can also run this step directly with [Bitrise CLI](https://github.com/bitrise-io/bitrise).

## ⚙️ Configuration

<details>
<summary>Inputs</summary>

| Key | Description | Flags | Default |
| --- | --- | --- | --- |
| `aab_path` | The **Android App Bundle path** input field is automatically filled out by the output of the previous build Step | required | `$BITRISE_AAB_PATH` |
| `keystore_url` | The keystore file's URL which is generated when you upload the file to the Code Signing tab. | required, sensitive | `$BITRISEIO_ANDROID_KEYSTORE_URL` |
| `keystore_password` | The password you added to the keystore. | required, sensitive | `$BITRISEIO_ANDROID_KEYSTORE_PASSWORD` |
| `keystore_alias` | Identifier name you added to the keystore. | required, sensitive | `$BITRISEIO_ANDROID_KEYSTORE_ALIAS` |
| `private_key_password` | Password you added to the private key. | sensitive | `$BITRISEIO_ANDROID_KEYSTORE_PRIVATE_KEY_PASSWORD` |
| `bundletool_version` | If you wish to set a specific version, add it here based on [Bundletool's official release](https://github.com/google/bundletool/releases) page. |  | `1.8.1` |
</details>

<details>
<summary>Outputs</summary>

| Environment Variable | Description |
| --- | --- |
| `BITRISE_APK_PATH` | The APK is exported to this output Environment Variable and can be picked up by the next Step or Ship. |
</details>

## 🙋 Contributing

We welcome [pull requests](https://github.com/bitrise-steplib/bitrise-step-export-universal-apk/pulls) and [issues](https://github.com/bitrise-steplib/bitrise-step-export-universal-apk/issues) against this repository.

For pull requests, work on your changes in a forked repository and use the Bitrise CLI to [run step tests locally](https://docs.bitrise.io/en/bitrise-ci/bitrise-cli/running-your-first-local-build-with-the-cli.html).

Learn more about developing steps:

- [Create your own step](https://docs.bitrise.io/en/bitrise-ci/workflows-and-pipelines/developing-your-own-bitrise-step/developing-a-new-step.html)
