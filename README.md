# Thumb Drive Commander 

[![Linux & OSX Build Status](https://travis-ci.org/xiegeo/thuder.svg?branch=master)](https://travis-ci.org/xiegeo/thuder)
[![Windows Build status](https://ci.appveyor.com/api/projects/status/bunpw2d87nm0vks5?svg=true)](https://ci.appveyor.com/project/xiegeo/thuder)



Thumb Drive Commander (thuder) is a tool that can (working but with incomplete features) push and pull data from a headless device with no internet access. 

## Usage:

- USB Thumb Drive (aka: sneakernet) for data transport (with FAT32 support for compatibility)
- As a service that waits for drive insertions (mounting can be done by USBmount) and does pre-configured actions which maybe supplemented by a settings file in the thumb drive.
- Different hosts (by hostname and hardware (cpu, microSD) IDs) uses different sub-folders in the same removable media.
- Preformed actions can be controlled by some authorization method that verifies the drive or its contents.
- After file transfer actions completes, thuder can run a list of commands, to tell relevant services to reload configurations
- After actions are finished and the drive flashed. The led on the thumb drive stops flashing. It is then safe to remove the drive.

- Alternatively thuder can be used as a library to easily filter and copy files recursively. 

## Why?

I need to remotely service devices with someone on location. The devices should stay in operation during service. The location might not have internet access, including cellular. I need to read logs and push fixes (and run scripts to reload changed settings and executables). Plug in an usb thumb drive seems like the most straightforward solution.


## Design Considerations:

### Security 

Security concerns if a removable media or data on such media is authorized. 

Since typical thumb drives (and many host devices such as rpi) have no build in security architecture, such as a trusted computing platform, this makes any addon to authenticate a device less than perfect. This is mitigated by the requirement that any attacker must be physically present to gain a point of entry. Preventing a worm can only be accomplished by limiting scripting capabilities of thuder or performing good hygiene when reusing thumb drives.

Authentication of data is also possible, but limit the environment where data can be modified.

The sample app requires a check for a specially crafted file in the removable media to allow any operations. You must define the contents of this file for the sample app to work

### Idempotent operation

Repeats of actions are no-ops. Tasks are performed similar to ansible, with a stronger emphasis on portability and fast file manipulations.

### Many to many relationship between hosts and removable media

Multiple hosts can use the same media by using separate sub directories. Multiple media can be used on the same host, where only the current connected storage is considered the most up to date. 

### Portability

#### File Permissions

(TODO)

The target permissions after written on the host file system are provided by meta-date in configuration files or by appending the octal flags in filenames for convenience.

The default permissions are 0755 for both files and folders of the user running the service (root).

For consistency, any permissions of existing files are ignored when setting the new permissions.

Pulling files does (may?) not retain permissions.

#### Case Sensitivity

There should not be a valid use case for the same filename of different cases to exist. Such directory structure can not be used on Windows.

When pushing files to the host, all files with other cases are deleted.

When deleting files, all files ignoring case are deleted.

When pulling files to the removable media, repeated file names in other cases are renamed so they are different files to case insensitive systems. (future)

#### Others

(TODO, only logs and continues to the next file for now)

Other portability issues, such as special characters, name length, file size, or anything else that produce an error will be logged and skipped (for pull/backup) or terminated (for push/update). More workarounds can be considered in the future. 

### Tasks and configuration

The modal that Ansible uses is an inspiration, but too complex for the current stage of development. Getting the core functionality to work comes first.