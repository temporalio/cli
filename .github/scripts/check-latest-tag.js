const { execFileSync } = require('child_process');
const fs = require('fs');

const releaseTag = process.env.RELEASE_TAG;
const version = releaseTag.startsWith('v') ? releaseTag.slice(1) : releaseTag;
const sourceImage = `temporalio/temporal:${version}`;
const latestImage = 'temporalio/temporal:latest';

// Set outputs for use in next step
fs.appendFileSync(process.env.GITHUB_OUTPUT, `version=${version}\n`);
fs.appendFileSync(process.env.GITHUB_OUTPUT, `source_image=${sourceImage}\n`);
fs.appendFileSync(process.env.GITHUB_OUTPUT, `latest_image=${latestImage}\n`);

console.log(`Version: ${version}`);
console.log('Checking if image is already tagged as latest...');

try {
  const latestManifest = execFileSync(
    'docker',
    ['manifest', 'inspect', latestImage],
    { encoding: 'utf8' }
  );
  const latestDigest = JSON.parse(latestManifest).config.digest;

  const sourceManifest = execFileSync(
    'docker',
    ['manifest', 'inspect', sourceImage],
    { encoding: 'utf8' }
  );
  const sourceDigest = JSON.parse(sourceManifest).config.digest;

  if (latestDigest === sourceDigest) {
    console.log(`Image ${version} is already tagged as latest`);
    fs.appendFileSync(process.env.GITHUB_OUTPUT, 'already_latest=true\n');
    process.exit(0);
  }

  console.log(`Latest digest: ${latestDigest}`);
  console.log(`Source digest: ${sourceDigest}`);
  console.log('Digests do not match, will update latest tag');
  fs.appendFileSync(process.env.GITHUB_OUTPUT, 'already_latest=false\n');

} catch (error) {
  console.log('Could not compare digests (image may not exist yet)');
  console.log(`Error: ${error.message}`);
  fs.appendFileSync(process.env.GITHUB_OUTPUT, 'already_latest=false\n');
}
