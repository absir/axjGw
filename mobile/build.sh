#!/usr/bin/env bash
cd `dirname $0`

export ANDROID_HOME=~/Library/Android/sdk/
export ANDROID_NDK_HOME=$ANDROID_HOME/ndk
# ln -s ndk/23.1.7779620  ndk-bundle

echo "IOS SDK BUILD...."
gomobile bind -tags=gomobile -target=ios  ./../src/pkg/asdk

echo "ANDROID SDK BUILD...."
gomobile bind -tags=gomobile -target=android  ./../src/pkg/asdk