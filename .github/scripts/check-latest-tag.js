const { execFileSync } = require('child_process');

// Get inputs from environment
const releaseTag = process.env.RELEASE_TAG;

// Strip 'v' prefix from version
const version = releaseTag.startsWith('v') ? releaseTag.slice(1) : releaseTag;
const sourceImage = `temporalio/temporal:${version}`;
const latestImage = 'temporalio/temporal:latest';

// Set outputs for use in next step
console.log(`::set-output name=version::${version}`);
console.log(`::set-output name=source_image::${sourceImage}`);
console.log(`::set-output name=latest_image::${latestImage}`);

console.log(`Version: ${version}`);
console.log('Checking if image is already tagged as latest...');

try {
  // Get digest for the latest tag
  const latestManifest = execFileSync(
    'docker',
    ['manifest', 'inspect', latestImage],
    { encoding: 'utf8' }
  );
  const latestDigest = JSON.parse(latestManifest).config.digest;

  // Get digest for the version tag
  const sourceManifest = execFileSync(
    'docker',
    ['manifest', 'inspect', sourceImage],
    { encoding: 'utf8' }
  );
  const sourceDigest = JSON.parse(sourceManifest).config.digest;

  // Compare digests
  if (latestDigest === sourceDigest) {
    console.log(`✅ Image ${version} is already tagged as latest`);
    console.log('::set-output name=already_latest::true');
    process.exit(0);
  }

  console.log(`Latest digest: ${latestDigest}`);
  console.log(`Source digest: ${sourceDigest}`);
  console.log('Digests do not match, will update latest tag');
  console.log('::set-output name=already_latest::false');

} catch (error) {
  console.log('Could not compare digests (image may not exist yet)');
  console.log(`Error: ${error.message}`);
  console.log('::set-output name=already_latest::false');
}
