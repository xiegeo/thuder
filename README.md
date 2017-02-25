# Thumb Drive Commander 

Thumb Drive Commander (thuder) is a tool that can (will) push and pull data from a headless device with no internet access. 

## Usage:

- USB Thumb Drive (aka: sneakernet) for data transport (with FAT32 support for compatibility)
- A service running on the Pi that waits for drive insertions (mounting can be done by USBmount) and does pre-configered actions which maybe overridden by a settings file in the thumb drive.
- Different Pis (by hostname and hardware (cpu, micoSD) IDs) can use different sub-folders in the thumb drive.
- Preformed actions can be controlled by some authorization method that verifies the drive or its contents.
- When actions are finished and the drive unmounted. The led on the thumb drive stops flashing. It is then safe to remove the drive.

## Why?

I need to remotely service devices with someone on location. The devices should stay in operation during service. The location might not have internet access, including cellular. I need to read logs and push fixes (and run scripts to reload changed settings and executables). Plug in an usb thumb drive seems like the most straightforward solution.


## Design Considerations:

### Security 

Security concerns if a removable media or data on such media is authorized. 

Since typical thumb drives have no build security achecture, such as a trusted computing platform, this makes any addon to authenticate a device less than perfect. This is mitigated by the requirement that any attacker must be phyically present to gain a point of entry. Preventing a worm can only be accomplished by limiting scripting capiblities of thuder or performing good hygiene when reusing thumb drives.

Authentication of data is also possible, but limit the environment where data can be modified.

### Idempotentcy

Repeat of the actions should be a noop. Tasks are performed simular to ansible, with a stronger emphasis on protablity and file manipulations 

### Many to many relationship between hosts and removable media

Multiple hosts can use the same media by using unque sub directories. Mutiple media can be used on the same host, where only the current connected storge is considered the most upto date. 

### Portablity

#### File Permissions

The target permissions after written on the host file system are provided by metadate in configeration files or by appending the octal flags in filenames for convinice.

The default permissions are 0755 for both files and folders of the user running the service (root).

For consistency, any permissions of existing files are ignored when setting the new permissions.

Pulling files does (may?) not retain permissions.

#### Case Sensitivity

There should not be a valid usecase for the same filename of different cases to exist. Such directory structure can not be used on Windows.

When pushing files to the host, all files with other cases are deleted.

When deleting files, all files ignoring case are deleted.

When pulling files to the removable media, repeated filesnames in other cases are renamed so they are different files to case insensitive systems. 

#### Others

Other portablity issues, such as special charcters, name length, file size, or anything else that produce an error will be logged and skipped (for pull/backup) or termiate (for push/update). More special case handling can be considered in the future. 

### Tasks and configeration

The modal that Ansible uses is an inspiration, but too complex for the current stage of development. Getting the core functionality to work comes first.