# export ANDROID_HOME=~/Library/Android/sdk/
# export ANDROID_NDK_HOME=$ANDROID_HOME/ndk
# ln -s ndk/23.1.7779620  ndk-bundle

echo "IOS SDK BUILD...."
gomobile bind -target=ios  ./../src/pkg/asdk 

echo "ANDROID SDK BUILD...."
gomobile bind -target=android  ./../src/pkg/asdk