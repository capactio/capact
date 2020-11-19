# Hub

The directory contains a representation of the OCH repository content.

The Hub contains the following groups:
- [`core`](./core) - Contains OCH metadata with configuration and core manifests, such as Interfaces or Types.
- [`implementation`](./implementation) - Contains Implementation manifests. Implementation describes an action and its prerequisites. Action implements a given interface. The runner handles the execution of the action according to the “action” property, which describes how the function is run.
- [`interface`](./interface) - Contains InterfaceGroup and Interface manifests. InterfaceGroup holds metadata for a group of Interfaces. Interface defines an action signature. It describes the action name, input and output parameters.
- [`type`](./type) - Contains Type manifests. Type is a primitive, which is used in Interfaces as a description and validation of input and output parameters.
- [`tag`](./tag) - Contains Tag manifests. Tag is used to categorize Implementations and Types.
- [`vendor`](./vendor) - Contains metadata for external OCH repositories that are connected as a part of federation support. 
