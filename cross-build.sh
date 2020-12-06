#!/usr/bin/env bash

# Adapted from https://gist.github.com/makeworld-the-better-one/e1bb127979ae4195f43aaa3ad46b1097/e8abbae0ce5af35a227ffaeb1c589ac071738fd2

type setopt >/dev/null 2>&1

NOT_ALLOWED_OS="windows js android ios illumos aix"
NOT_ALLOWED_ARCH="mips mips64 mips64le mipsle ppc64 ppc64le riscv64 s390x"

FAILURES=""
BASE_DIR="$(pwd)"

contains() {
    # Source: https://stackoverflow.com/a/8063398/7361270
    [[ $1 =~ (^|[[:space:]])$2($|[[:space:]]) ]]
}

export GO111MODULES=on
export CGO_ENABLED=0

rm -rf build
mkdir build

# Get all targets
while IFS= read -r target; do
    GOOS=${target%/*}
    GOARCH=${target#*/}
    
    if contains "$NOT_ALLOWED_OS" "$GOOS" ; then
        continue
    fi
    if contains "$NOT_ALLOWED_ARCH" "$GOARCH" ; then
        continue
    fi

    if [[ $GOOS == "darwin" ]] && [[ $GOARCH = "arm64" ]]; then
        continue
    fi
    
    for binary in view like add-comment; do
        BIN_FILENAME="$binary"
        cd "$BASE_DIR/$binary"

        # Check for arm and set arm version
        if [[ $GOARCH == "arm" ]]; then
            # Set what arm versions each platform supports
            if [[ $GOOS == "darwin" ]]; then
                arms="7"
            elif [[ $GOOS == "windows" ]]; then
                # This is a guess, it's not clear what Windows supports from the docs
                # But I was able to build all these on my machine
                arms="5 6 7" 
            elif [[ $GOOS == *"bsd" ]]; then
                arms="6 7"
            else
                # Linux goes here
                arms="5 6 7"
            fi

            # Now do the arm build
            for GOARM in $arms; do
                if [[ "${GOOS}" == "windows" ]]; then BIN_FILENAME="${BIN_FILENAME}.exe"; fi
                CMD="GOARM=${GOARM} GOOS=${GOOS} GOARCH=${GOARCH} go build -o ${BIN_FILENAME} $@"
                echo "${CMD}"
                eval "${CMD}"
                status=$?
                if [ $status -eq 0 ]; then
                    # Move binarIES
                    mkdir -p "../build/$GOOS-arm$GOARM"
                    mv "$BIN_FILENAME" "../build/$GOOS-arm$GOARM"
                else
                    FAILURES="${FAILURES} ${GOOS}/${GOARCH}${GOARM}"
                fi
            done

            continue  # Skip the non-arm building done below
        fi

        # Build non-arm here
        if [[ "${GOOS}" == "windows" ]]; then BIN_FILENAME="${BIN_FILENAME}.exe"; fi
        CMD="GOOS=${GOOS} GOARCH=${GOARCH} go build -o ${BIN_FILENAME} $@"
        echo "${CMD}"
        eval "${CMD}"
        status=$?
        if [ $status -eq 0 ]; then
            # Move binary
            mkdir -p "../build/$GOOS-$GOARCH"
            mv "$BIN_FILENAME" "../build/$GOOS-$GOARCH"
        else
            FAILURES="${FAILURES} ${GOOS}/${GOARCH}"
        fi

    done
done <<< "$(go tool dist list)"

# Create .tar.gz of each folder
cd "$BASE_DIR/build"
echo "Creating tar archives..."

for dir in */; do
    dir=${dir%*/}  # remove the trailing "/"
    tar czf "$dir.tar.gz" "$dir"
    rm -r "$dir"
done
echo "Done all."

if [[ "${FAILURES}" != "" ]]; then
    echo ""
    echo "${SCRIPT_NAME} failed on: ${FAILURES}"
    exit 1
fi
