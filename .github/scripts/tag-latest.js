const { execFileSync } = require('child_process');

const inputTag = process.env.INPUT_TAG;
const inputSha = process.env.INPUT_SHA;

// Validate inputs
if (!inputTag && !inputSha) {
  console.error('Error: Either "tag" or "sha" input must be provided');
  process.exit(1);
}

if (inputTag && inputSha) {
  console.error('Error: Only one of "tag" or "sha" should be provided, not both');
  process.exit(1);
}

// Determine source image
let sourceImage;
let imageRef;

if (inputTag) {
  const version = inputTag.startsWith('v') ? inputTag.slice(1) : inputTag;
  sourceImage = `temporalio/temporal:${version}`;
  imageRef = version;
  console.log(`Using tag: ${version}`);
} else {
  if (!inputSha.startsWith('sha256:')) {
    console.error('Error: SHA must start with "sha256:"');
    process.exit(1);
  }
  sourceImage = `temporalio/temporal@${inputSha}`;
  imageRef = inputSha;
  console.log(`Using SHA: ${inputSha}`);
}

const latestImage = 'temporalio/temporal:latest';

// Check if already tagged as latest
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
    console.log(`ℹ️ Image ${imageRef} is already tagged as latest. No action needed.`);
    process.exit(0);
  }

  console.log(`Latest digest: ${latestDigest}`);
  console.log(`Source digest: ${sourceDigest}`);
} catch (error) {
  console.log('Could not compare digests, will proceed with tagging');
}

// Pull, tag, and push
try {
  console.log(`Pulling ${sourceImage}...`);
  execFileSync('docker', ['pull', sourceImage], { stdio: 'inherit' });

  console.log(`Tagging as ${latestImage}...`);
  execFileSync('docker', ['tag', sourceImage, latestImage], { stdio: 'inherit' });

  console.log(`Pushing ${latestImage}...`);
  execFileSync('docker', ['push', latestImage], { stdio: 'inherit' });

  console.log(`✅ Successfully updated latest tag to point to ${imageRef}`);
} catch (error) {
  console.error('Error during Docker operations:', error.message);
  process.exit(1);
}
